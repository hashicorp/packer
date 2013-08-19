package virtualbox

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const BuilderId = "mitchellh.virtualbox"

type Builder struct {
	config config
	runner multistep.Runner
}

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	BootCommand          []string   `mapstructure:"boot_command"`
	DiskSize             uint       `mapstructure:"disk_size"`
	FloppyFiles          []string   `mapstructure:"floppy_files"`
	GuestAdditionsPath   string     `mapstructure:"guest_additions_path"`
	GuestAdditionsURL    string     `mapstructure:"guest_additions_url"`
	GuestAdditionsSHA256 string     `mapstructure:"guest_additions_sha256"`
	GuestOSType          string     `mapstructure:"guest_os_type"`
	Headless             bool       `mapstructure:"headless"`
	HTTPDir              string     `mapstructure:"http_directory"`
	HTTPPortMin          uint       `mapstructure:"http_port_min"`
	HTTPPortMax          uint       `mapstructure:"http_port_max"`
	ISOChecksum          string     `mapstructure:"iso_checksum"`
	ISOChecksumType      string     `mapstructure:"iso_checksum_type"`
	ISOUrls              []string   `mapstructure:"iso_urls"`
	OutputDir            string     `mapstructure:"output_directory"`
	ShutdownCommand      string     `mapstructure:"shutdown_command"`
	SSHHostPortMin       uint       `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax       uint       `mapstructure:"ssh_host_port_max"`
	SSHPassword          string     `mapstructure:"ssh_password"`
	SSHPort              uint       `mapstructure:"ssh_port"`
	SSHUser              string     `mapstructure:"ssh_username"`
	VBoxVersionFile      string     `mapstructure:"virtualbox_version_file"`
	VBoxManage           [][]string `mapstructure:"vboxmanage"`
	VMName               string     `mapstructure:"vm_name"`
	Format         	     string     `mapstructure:"format"`

	RawBootWait        string `mapstructure:"boot_wait"`
	RawSingleISOUrl    string `mapstructure:"iso_url"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`
	RawSSHWaitTimeout  string `mapstructure:"ssh_wait_timeout"`

	bootWait        time.Duration ``
	shutdownTimeout time.Duration ``
	sshWaitTimeout  time.Duration ``
	tpl             *packer.ConfigTemplate
}

func (b *Builder) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.FloppyFiles == nil {
		b.config.FloppyFiles = make([]string, 0)
	}

	if b.config.GuestAdditionsPath == "" {
		b.config.GuestAdditionsPath = "VBoxGuestAdditions.iso"
	}

	if b.config.GuestOSType == "" {
		b.config.GuestOSType = "Other"
	}

	if b.config.HTTPPortMin == 0 {
		b.config.HTTPPortMin = 8000
	}

	if b.config.HTTPPortMax == 0 {
		b.config.HTTPPortMax = 9000
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}

	if b.config.RawBootWait == "" {
		b.config.RawBootWait = "10s"
	}

	if b.config.SSHHostPortMin == 0 {
		b.config.SSHHostPortMin = 2222
	}

	if b.config.SSHHostPortMax == 0 {
		b.config.SSHHostPortMax = 4444
	}

	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}

	if b.config.VBoxManage == nil {
		b.config.VBoxManage = make([][]string, 0)
	}

	if b.config.VBoxVersionFile == "" {
		b.config.VBoxVersionFile = ".vbox_version"
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	if b.config.Format == "" {
		b.config.Format = "ovf"
	}

	// Errors
	templates := map[string]*string{
		"guest_additions_sha256":  &b.config.GuestAdditionsSHA256,
		"guest_os_type":           &b.config.GuestOSType,
		"http_directory":          &b.config.HTTPDir,
		"iso_checksum":            &b.config.ISOChecksum,
		"iso_checksum_type":       &b.config.ISOChecksumType,
		"iso_url":                 &b.config.RawSingleISOUrl,
		"output_directory":        &b.config.OutputDir,
		"shutdown_command":        &b.config.ShutdownCommand,
		"ssh_password":            &b.config.SSHPassword,
		"ssh_username":            &b.config.SSHUser,
		"virtualbox_version_file": &b.config.VBoxVersionFile,
		"vm_name":                 &b.config.VMName,
		"format":                  &b.config.Format,
		"boot_wait":               &b.config.RawBootWait,
		"shutdown_timeout":        &b.config.RawShutdownTimeout,
		"ssh_wait_timeout":        &b.config.RawSSHWaitTimeout,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	for i, url := range b.config.ISOUrls {
		var err error
		b.config.ISOUrls[i], err = b.config.tpl.Process(url, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing iso_urls[%d]: %s", i, err))
		}
	}

	validates := map[string]*string{
		"guest_additions_path": &b.config.GuestAdditionsPath,
		"guest_additions_url":  &b.config.GuestAdditionsURL,
	}

	for n, ptr := range validates {
		if err := b.config.tpl.Validate(*ptr); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing %s: %s", n, err))
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

	if !(b.config.Format == "ovf" || b.config.Format == "ova") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("invalid format, only 'ovf' or 'ova' are allowed"))
	}

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

	if b.config.RawSingleISOUrl == "" && len(b.config.ISOUrls) == 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("One of iso_url or iso_urls must be specified."))
	} else if b.config.RawSingleISOUrl != "" && len(b.config.ISOUrls) > 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Only one of iso_url or iso_urls may be specified."))
	} else if b.config.RawSingleISOUrl != "" {
		b.config.ISOUrls = []string{b.config.RawSingleISOUrl}
	}

	for i, url := range b.config.ISOUrls {
		b.config.ISOUrls[i], err = common.DownloadableURL(url)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Failed to parse iso_url %d: %s", i+1, err))
		}
	}

	if b.config.GuestAdditionsSHA256 != "" {
		b.config.GuestAdditionsSHA256 = strings.ToLower(b.config.GuestAdditionsSHA256)
	}

	if !b.config.PackerForce {
		if _, err := os.Stat(b.config.OutputDir); err == nil {
			errs = packer.MultiErrorAppend(
				errs,
				fmt.Errorf("Output directory '%s' already exists. It must not exist.", b.config.OutputDir))
		}
	}

	b.config.bootWait, err = time.ParseDuration(b.config.RawBootWait)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
	}

	if b.config.RawShutdownTimeout == "" {
		b.config.RawShutdownTimeout = "5m"
	}

	if b.config.RawSSHWaitTimeout == "" {
		b.config.RawSSHWaitTimeout = "20m"
	}

	b.config.shutdownTimeout, err = time.ParseDuration(b.config.RawShutdownTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	if b.config.SSHHostPortMin > b.config.SSHHostPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
	}

	if b.config.SSHUser == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("An ssh_username must be specified."))
	}

	b.config.sshWaitTimeout, err = time.ParseDuration(b.config.RawSSHWaitTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed parsing ssh_wait_timeout: %s", err))
	}

	for i, args := range b.config.VBoxManage {
		for j, arg := range args {
			if err := b.config.tpl.Validate(arg); err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("Error processing vboxmanage[%d][%d]: %s", i, j, err))
			}
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with VirtualBox
	driver, err := b.newDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating VirtualBox driver: %s", err)
	}

	steps := []multistep.Step{
		new(stepDownloadGuestAdditions),
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			ResultKey:    "iso_path",
			Url:          b.config.ISOUrls,
		},
		new(stepPrepareOutputDir),
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		new(stepHTTPServer),
		new(stepSuppressMessages),
		new(stepCreateVM),
		new(stepCreateDisk),
		new(stepAttachISO),
		new(stepAttachFloppy),
		new(stepForwardSSH),
		new(stepVBoxManage),
		new(stepRun),
		new(stepTypeBootCommand),
		&common.StepConnectSSH{
			SSHAddress:     sshAddress,
			SSHConfig:      sshConfig,
			SSHWaitTimeout: b.config.sshWaitTimeout,
		},
		new(stepUploadVersion),
		new(stepUploadGuestAdditions),
		new(common.StepProvision),
		new(stepShutdown),
		new(stepExport),
	}

	// Setup the state bag
	state := make(map[string]interface{})
	state["cache"] = cache
	state["config"] = &b.config
	state["driver"] = driver
	state["hook"] = hook
	state["ui"] = ui

	// Run
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
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}

		return err
	}

	if err := filepath.Walk(b.config.OutputDir, visit); err != nil {
		return nil, err
	}

	artifact := &Artifact{
		dir: b.config.OutputDir,
		f:   files,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder) newDriver() (Driver, error) {
	vboxmanagePath, err := exec.LookPath("VBoxManage")
	if err != nil {
		return nil, err
	}

	log.Printf("VBoxManage path: %s", vboxmanagePath)
	driver := &VBox42Driver{vboxmanagePath}
	if err := driver.Verify(); err != nil {
		return nil, err
	}

	return driver, nil
}
