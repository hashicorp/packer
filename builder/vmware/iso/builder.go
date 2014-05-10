package iso

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

const BuilderIdESX = "mitchellh.vmware-esx"

type Builder struct {
	config config
	runner multistep.Runner
}

type config struct {
	common.PackerConfig      `mapstructure:",squash"`
	vmwcommon.DriverConfig   `mapstructure:",squash"`
	vmwcommon.OutputConfig   `mapstructure:",squash"`
	vmwcommon.RunConfig      `mapstructure:",squash"`
	vmwcommon.ShutdownConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig      `mapstructure:",squash"`
	vmwcommon.ToolsConfig    `mapstructure:",squash"`
	vmwcommon.VMXConfig      `mapstructure:",squash"`

	DiskName        string   `mapstructure:"vmdk_name"`
	DiskSize        uint     `mapstructure:"disk_size"`
	DiskTypeId      string   `mapstructure:"disk_type_id"`
	FloppyFiles     []string `mapstructure:"floppy_files"`
	GuestOSType     string   `mapstructure:"guest_os_type"`
	ISOChecksum     string   `mapstructure:"iso_checksum"`
	ISOChecksumType string   `mapstructure:"iso_checksum_type"`
	ISOUrls         []string `mapstructure:"iso_urls"`
	VMName          string   `mapstructure:"vm_name"`
	HTTPDir         string   `mapstructure:"http_directory"`
	HTTPPortMin     uint     `mapstructure:"http_port_min"`
	HTTPPortMax     uint     `mapstructure:"http_port_max"`
	BootCommand     []string `mapstructure:"boot_command"`
	SkipCompaction  bool     `mapstructure:"skip_compaction"`
	VMXTemplatePath string   `mapstructure:"vmx_template_path"`
	VNCPortMin      uint     `mapstructure:"vnc_port_min"`
	VNCPortMax      uint     `mapstructure:"vnc_port_max"`

	RemoteType      string `mapstructure:"remote_type"`
	RemoteDatastore string `mapstructure:"remote_datastore"`
	RemoteHost      string `mapstructure:"remote_host"`
	RemotePort      uint   `mapstructure:"remote_port"`
	RemoteUser      string `mapstructure:"remote_username"`
	RemotePassword  string `mapstructure:"remote_password"`

	RawSingleISOUrl string `mapstructure:"iso_url"`

	tpl *packer.ConfigTemplate
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.DriverConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.OutputConfig.Prepare(b.config.tpl, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.ToolsConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.VMXConfig.Prepare(b.config.tpl)...)
	warnings := make([]string, 0)

	if b.config.DiskName == "" {
		b.config.DiskName = "disk"
	}

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.DiskTypeId == "" {
		// Default is growable virtual disk split in 2GB files.
		b.config.DiskTypeId = "1"

		if b.config.RemoteType == "esx5" {
			b.config.DiskTypeId = "zeroedthick"
		}
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

	if b.config.VNCPortMin == 0 {
		b.config.VNCPortMin = 5900
	}

	if b.config.VNCPortMax == 0 {
		b.config.VNCPortMax = 6000
	}

	if b.config.RemoteUser == "" {
		b.config.RemoteUser = "root"
	}

	if b.config.RemoteDatastore == "" {
		b.config.RemoteDatastore = "datastore1"
	}

	if b.config.RemotePort == 0 {
		b.config.RemotePort = 22
	}

	// Errors
	templates := map[string]*string{
		"disk_name":         &b.config.DiskName,
		"guest_os_type":     &b.config.GuestOSType,
		"http_directory":    &b.config.HTTPDir,
		"iso_checksum":      &b.config.ISOChecksum,
		"iso_checksum_type": &b.config.ISOChecksumType,
		"iso_url":           &b.config.RawSingleISOUrl,
		"vm_name":           &b.config.VMName,
		"vmx_template_path": &b.config.VMXTemplatePath,
		"remote_type":       &b.config.RemoteType,
		"remote_host":       &b.config.RemoteHost,
		"remote_datastore":  &b.config.RemoteDatastore,
		"remote_user":       &b.config.RemoteUser,
		"remote_password":   &b.config.RemotePassword,
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

	if b.config.HTTPPortMin > b.config.HTTPPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("http_port_min must be less than http_port_max"))
	}

	if b.config.ISOChecksumType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("The iso_checksum_type must be specified."))
	} else {
		b.config.ISOChecksumType = strings.ToLower(b.config.ISOChecksumType)
		if b.config.ISOChecksumType != "none" {
			if b.config.ISOChecksum == "" {
				errs = packer.MultiErrorAppend(
					errs, errors.New("Due to large file sizes, an iso_checksum is required"))
			} else {
				b.config.ISOChecksum = strings.ToLower(b.config.ISOChecksum)
			}

			if h := common.HashForType(b.config.ISOChecksumType); h == nil {
				errs = packer.MultiErrorAppend(
					errs,
					fmt.Errorf("Unsupported checksum type: %s", b.config.ISOChecksumType))
			}
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

	if b.config.VMXTemplatePath != "" {
		if err := b.validateVMXTemplatePath(); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("vmx_template_path is invalid: %s", err))
		}

	}

	if b.config.VNCPortMin > b.config.VNCPortMax {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	// Remote configuration validation
	if b.config.RemoteType != "" {
		if b.config.RemoteHost == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("remote_host must be specified"))
		}
	}

	// Warnings
	if b.config.ISOChecksumType == "none" {
		warnings = append(warnings,
			"A checksum type of 'none' was specified. Since ISO files are so big,\n"+
				"a checksum is highly recommended.")
	}

	if b.config.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	driver, err := NewDriver(&b.config)
	if err != nil {
		return nil, fmt.Errorf("Failed creating VMware driver: %s", err)
	}

	// Determine the output dir implementation
	var dir OutputDir
	switch d := driver.(type) {
	case OutputDir:
		dir = d
	default:
		dir = new(vmwcommon.LocalOutputDir)
	}
	dir.SetOutputDir(b.config.OutputDir)

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("dir", dir)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	steps := []multistep.Step{
		&vmwcommon.StepPrepareTools{
			RemoteType:        b.config.RemoteType,
			ToolsUploadFlavor: b.config.ToolsUploadFlavor,
		},
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			ResultKey:    "iso_path",
			Url:          b.config.ISOUrls,
		},
		&vmwcommon.StepOutputDir{
			Force: b.config.PackerForce,
		},
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		&stepRemoteUpload{
			Key:     "floppy_path",
			Message: "Uploading Floppy to remote machine...",
		},
		&stepRemoteUpload{
			Key:     "iso_path",
			Message: "Uploading ISO to remote machine...",
		},
		&stepCreateDisk{},
		&stepCreateVMX{},
		&vmwcommon.StepConfigureVMX{
			CustomData: b.config.VMXData,
		},
		&vmwcommon.StepSuppressMessages{},
		&stepHTTPServer{},
		&stepConfigureVNC{},
		&StepRegister{},
		&vmwcommon.StepRun{
			BootWait:           b.config.BootWait,
			DurationBeforeStop: 5 * time.Second,
			Headless:           b.config.Headless,
		},
		&stepTypeBootCommand{},
		&common.StepConnectSSH{
			SSHAddress:     driver.SSHAddress,
			SSHConfig:      vmwcommon.SSHConfigFunc(&b.config.SSHConfig),
			SSHWaitTimeout: b.config.SSHWaitTimeout,
			NoPty:          b.config.SSHSkipRequestPty,
		},
		&vmwcommon.StepUploadTools{
			RemoteType:        b.config.RemoteType,
			ToolsUploadFlavor: b.config.ToolsUploadFlavor,
			ToolsUploadPath:   b.config.ToolsUploadPath,
			Tpl:               b.config.tpl,
		},
		&common.StepProvision{},
		&vmwcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&vmwcommon.StepCleanFiles{},
		&vmwcommon.StepCleanVMX{},
		&vmwcommon.StepConfigureVMX{
			CustomData: b.config.VMXDataPost,
		},
		&vmwcommon.StepCompactDisk{
			Skip: b.config.SkipCompaction,
		},
	}

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
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	// Compile the artifact list
	files, err := state.Get("dir").(OutputDir).ListFiles()
	if err != nil {
		return nil, err
	}

	// Set the proper builder ID
	builderId := vmwcommon.BuilderId
	if b.config.RemoteType != "" {
		builderId = BuilderIdESX
	}

	return &Artifact{
		builderId: builderId,
		dir:       dir,
		f:         files,
	}, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder) validateVMXTemplatePath() error {
	f, err := os.Open(b.config.VMXTemplatePath)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	return b.config.tpl.Validate(string(data))
}
