package vmware

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const BuilderId = "mitchellh.vmware"

type Builder struct {
	config config
	driver Driver
	runner multistep.Runner
}

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	DiskName          string            `mapstructure:"vmdk_name"`
	DiskSize          uint              `mapstructure:"disk_size"`
	SingleDisk        bool              `mapstructure:"single_disk"`
	FloppyFiles       []string          `mapstructure:"floppy_files"`
	GuestOSType       string            `mapstructure:"guest_os_type"`
	ISOChecksum       string            `mapstructure:"iso_checksum"`
	ISOChecksumType   string            `mapstructure:"iso_checksum_type"`
	ISOUrl            string            `mapstructure:"iso_url"`
	VMName            string            `mapstructure:"vm_name"`
	OutputDir         string            `mapstructure:"output_directory"`
	Headless          bool              `mapstructure:"headless"`
	HTTPDir           string            `mapstructure:"http_directory"`
	HTTPPortMin       uint              `mapstructure:"http_port_min"`
	HTTPPortMax       uint              `mapstructure:"http_port_max"`
	BootCommand       []string          `mapstructure:"boot_command"`
	SkipCompaction    bool              `mapstructure:"skip_compaction"`
	SkipPtyRequest    bool              `mapstructure:"skip_pty_request"`
	ShutdownCommand   string            `mapstructure:"shutdown_command"`
	SSHUser           string            `mapstructure:"ssh_username"`
	SSHPassword       string            `mapstructure:"ssh_password"`
	SSHPort           uint              `mapstructure:"ssh_port"`
	ToolsUploadFlavor string            `mapstructure:"tools_upload_flavor"`
	ToolsUploadPath   string            `mapstructure:"tools_upload_path"`
	VMXData           map[string]string `mapstructure:"vmx_data"`
	VNCPortMin        uint              `mapstructure:"vnc_port_min"`
	VNCPortMax        uint              `mapstructure:"vnc_port_max"`

	RawBootWait        string `mapstructure:"boot_wait"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`
	RawSSHWaitTimeout  string `mapstructure:"ssh_wait_timeout"`

	bootWait        time.Duration ``
	shutdownTimeout time.Duration ``
	sshWaitTimeout  time.Duration ``
	tpl             *common.Template
}

func (b *Builder) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return err
	}

	b.config.tpl, err = common.NewTemplate()
	if err != nil {
		return err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if b.config.DiskName == "" {
		b.config.DiskName = "disk"
	}

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.FloppyFiles == nil {
		b.config.FloppyFiles = make([]string, 0)
	}

	if b.config.GuestOSType == "" {
		b.config.GuestOSType = "other"
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	if b.config.HTTPPortMin == 0 {
		b.config.HTTPPortMin = 8000
	}

	if b.config.HTTPPortMax == 0 {
		b.config.HTTPPortMax = 9000
	}

	if b.config.RawBootWait == "" {
		b.config.RawBootWait = "10s"
	}

	if b.config.VNCPortMin == 0 {
		b.config.VNCPortMin = 5900
	}

	if b.config.VNCPortMax == 0 {
		b.config.VNCPortMax = 6000
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}

	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}

	if b.config.ToolsUploadPath == "" {
		b.config.ToolsUploadPath = "{{ .Flavor }}.iso"
	}

	// Errors
	templates := map[string]*string{
		"disk_name":           &b.config.DiskName,
		"guest_os_type":       &b.config.GuestOSType,
		"http_directory":      &b.config.HTTPDir,
		"iso_checksum":        &b.config.ISOChecksum,
		"iso_checksum_type":   &b.config.ISOChecksumType,
		"iso_url":             &b.config.ISOUrl,
		"output_directory":    &b.config.OutputDir,
		"shutdown_command":    &b.config.ShutdownCommand,
		"ssh_password":        &b.config.SSHPassword,
		"ssh_username":        &b.config.SSHUser,
		"tools_upload_flavor": &b.config.ToolsUploadFlavor,
		"vm_name":             &b.config.VMName,
		"boot_wait":           &b.config.RawBootWait,
		"shutdown_timeout":    &b.config.RawShutdownTimeout,
		"ssh_wait_timeout":    &b.config.RawSSHWaitTimeout,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	for i, command := range b.config.BootCommand {
		if err := b.config.tpl.Validate(command); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing boot_command[%d]: %s", i, err))
		}
	}

	for i, file := range b.config.FloppyFiles {
		var err error
		b.config.FloppyFiles[i], err = b.config.tpl.Process(file, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing floppy_files[%d]: %s",
					i, err))
		}
	}

	newVMXData := make(map[string]string)
	for k, v := range b.config.VMXData {
		k, err = b.config.tpl.Process(k, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing VMX data key %s: %s", k, err))
			continue
		}

		v, err = b.config.tpl.Process(v, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing VMX data value '%s': %s", v, err))
			continue
		}

		newVMXData[k] = v
	}

	b.config.VMXData = newVMXData

	if b.config.HTTPPortMin > b.config.HTTPPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("http_port_min must be less than http_port_max"))
	}

	if b.config.ISOChecksum == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Due to large file sizes, an iso_checksum is required"))
	} else {
		b.config.ISOChecksum = strings.ToLower(b.config.ISOChecksum)
	}

	if b.config.ISOChecksumType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		b.config.ISOChecksumType = strings.ToLower(b.config.ISOChecksumType)
		if h := common.HashForType(b.config.ISOChecksumType); h == nil {
			errs = packer.MultiErrorAppend(
				errs,
				fmt.Errorf("Unsupported checksum type: %s", b.config.ISOChecksumType))
		}
	}

	if b.config.ISOUrl == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("An iso_url must be specified."))
	} else {
		b.config.ISOUrl, err = common.DownloadableURL(b.config.ISOUrl)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("iso_url: %s", err))
		}
	}

	if !b.config.PackerForce {
		if _, err := os.Stat(b.config.OutputDir); err == nil {
			errs = packer.MultiErrorAppend(
				errs,
				fmt.Errorf("Output directory '%s' already exists. It must not exist.", b.config.OutputDir))
		}
	}

	if b.config.SSHUser == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("An ssh_username must be specified."))
	}

	if b.config.RawBootWait != "" {
		b.config.bootWait, err = time.ParseDuration(b.config.RawBootWait)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
		}
	}

	if b.config.RawShutdownTimeout == "" {
		b.config.RawShutdownTimeout = "5m"
	}

	b.config.shutdownTimeout, err = time.ParseDuration(b.config.RawShutdownTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	if b.config.RawSSHWaitTimeout == "" {
		b.config.RawSSHWaitTimeout = "20m"
	}

	b.config.sshWaitTimeout, err = time.ParseDuration(b.config.RawSSHWaitTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_wait_timeout: %s", err))
	}

	if _, err := template.New("path").Parse(b.config.ToolsUploadPath); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("tools_upload_path invalid: %s", err))
	}

	if b.config.VNCPortMin > b.config.VNCPortMax {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	b.driver, err = NewDriver()
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed creating VMware driver: %s", err))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	steps := []multistep.Step{
		&stepPrepareTools{},
		&stepDownloadISO{},
		&stepPrepareOutputDir{},
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		&stepCreateDisk{},
		&stepCreateVMX{},
		&stepHTTPServer{},
		&stepConfigureVNC{},
		&stepRun{},
		&stepTypeBootCommand{},
		&common.StepConnectSSH{
			SkipPtyRequest: b.config.SkipPtyRequest,
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: b.config.sshWaitTimeout,
		},
		&stepUploadTools{},
		&common.StepProvision{},
		&stepShutdown{},
		&stepCleanFiles{},
		&stepCleanVMX{},
		&stepCompactDisk{},
	}

	// Setup the state bag
	state := make(map[string]interface{})
	state["cache"] = cache
	state["config"] = &b.config
	state["driver"] = b.driver
	state["hook"] = hook
	state["ui"] = ui

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state["error"]; ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state[multistep.StateCancelled]; ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state[multistep.StateHalted]; ok {
		return nil, errors.New("Build was halted.")
	}

	// Compile the artifact list
	files := make([]string, 0, 10)
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(b.config.OutputDir, visit); err != nil {
		return nil, err
	}

	return &Artifact{b.config.OutputDir, files}, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
