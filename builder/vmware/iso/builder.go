package iso

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderIdESX = "mitchellh.vmware-esx"

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	common.HTTPConfig        `mapstructure:",squash"`
	common.ISOConfig         `mapstructure:",squash"`
	vmwcommon.DriverConfig   `mapstructure:",squash"`
	vmwcommon.OutputConfig   `mapstructure:",squash"`
	vmwcommon.RunConfig      `mapstructure:",squash"`
	vmwcommon.ShutdownConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig      `mapstructure:",squash"`
	vmwcommon.ToolsConfig    `mapstructure:",squash"`
	vmwcommon.VMXConfig      `mapstructure:",squash"`

	AdditionalDiskSize  []uint   `mapstructure:"disk_additional_size"`
	DiskName            string   `mapstructure:"vmdk_name"`
	DiskSize            uint     `mapstructure:"disk_size"`
	DiskTypeId          string   `mapstructure:"disk_type_id"`
	FloppyFiles         []string `mapstructure:"floppy_files"`
	Format              string   `mapstructure:"format"`
	GuestOSType         string   `mapstructure:"guest_os_type"`
	Version             string   `mapstructure:"version"`
	VMName              string   `mapstructure:"vm_name"`
	BootCommand         []string `mapstructure:"boot_command"`
	KeepRegistered      bool     `mapstructure:"keep_registered"`
	SkipCompaction      bool     `mapstructure:"skip_compaction"`
	VMXTemplatePath     string   `mapstructure:"vmx_template_path"`
	VMXDiskTemplatePath string   `mapstructure:"vmx_disk_template_path"`

	RemoteType           string `mapstructure:"remote_type"`
	RemoteDatastore      string `mapstructure:"remote_datastore"`
	RemoteCacheDatastore string `mapstructure:"remote_cache_datastore"`
	RemoteCacheDirectory string `mapstructure:"remote_cache_directory"`
	RemoteHost           string `mapstructure:"remote_host"`
	RemotePort           uint   `mapstructure:"remote_port"`
	RemoteUser           string `mapstructure:"remote_username"`
	RemotePassword       string `mapstructure:"remote_password"`
	RemotePrivateKey     string `mapstructure:"remote_private_key_file"`

	ctx interpolate.Context
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"tools_upload_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	isoWarnings, isoErrs := b.config.ISOConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)
	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.DriverConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ToolsConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VMXConfig.Prepare(&b.config.ctx)...)

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

	if b.config.Version == "" {
		b.config.Version = "9"
	}

	if b.config.RemoteUser == "" {
		b.config.RemoteUser = "root"
	}

	if b.config.RemoteDatastore == "" {
		b.config.RemoteDatastore = "datastore1"
	}

	if b.config.RemoteCacheDatastore == "" {
		b.config.RemoteCacheDatastore = b.config.RemoteDatastore
	}

	if b.config.RemoteCacheDirectory == "" {
		b.config.RemoteCacheDirectory = "packer_cache"
	}

	if b.config.RemotePort == 0 {
		b.config.RemotePort = 22
	}
	if b.config.VMXTemplatePath != "" {
		if err := b.validateVMXTemplatePath(); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("vmx_template_path is invalid: %s", err))
		}

	}

	// Remote configuration validation
	if b.config.RemoteType != "" {
		if b.config.RemoteHost == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("remote_host must be specified"))
		}
	}

	if b.config.Format != "" {
		if !(b.config.Format == "ova" || b.config.Format == "ovf" || b.config.Format == "vmx") {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("format must be one of ova, ovf, or vmx"))
		}
	}

	// Warnings
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
	if b.config.RemoteType != "" && b.config.Format != "" {
		b.config.OutputDir = b.config.VMName
	}
	dir.SetOutputDir(b.config.OutputDir)

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("dir", dir)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&vmwcommon.StepPrepareTools{
			RemoteType:        b.config.RemoteType,
			ToolsUploadFlavor: b.config.ToolsUploadFlavor,
		},
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			Extension:    "iso",
			ResultKey:    "iso_path",
			TargetPath:   b.config.TargetPath,
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
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		&vmwcommon.StepConfigureVNC{
			VNCBindAddress: b.config.VNCBindAddress,
			VNCPortMin:     b.config.VNCPortMin,
			VNCPortMax:     b.config.VNCPortMax,
		},
		&StepRegister{
			Format: b.config.Format,
		},
		&vmwcommon.StepRun{
			BootWait:           b.config.BootWait,
			DurationBeforeStop: 5 * time.Second,
			Headless:           b.config.Headless,
		},
		&vmwcommon.StepTypeBootCommand{
			BootCommand: b.config.BootCommand,
			VMName:      b.config.VMName,
			Ctx:         b.config.ctx,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      driver.CommHost,
			SSHConfig: vmwcommon.SSHConfigFunc(&b.config.SSHConfig),
		},
		&vmwcommon.StepUploadTools{
			RemoteType:        b.config.RemoteType,
			ToolsUploadFlavor: b.config.ToolsUploadFlavor,
			ToolsUploadPath:   b.config.ToolsUploadPath,
			Ctx:               b.config.ctx,
		},
		&common.StepProvision{},
		&vmwcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&vmwcommon.StepCleanFiles{},
		&vmwcommon.StepCompactDisk{
			Skip: b.config.SkipCompaction,
		},
		&vmwcommon.StepConfigureVMX{
			CustomData: b.config.VMXDataPost,
			SkipFloppy: true,
		},
		&vmwcommon.StepCleanVMX{},
		&StepUploadVMX{
			RemoteType: b.config.RemoteType,
		},
		&StepExport{
			Format: b.config.Format,
		},
	}

	// Run!
	if b.config.PackerDebug {
		pauseFn := common.MultistepDebugFn(ui)
		state.Put("pauseFn", pauseFn)
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: pauseFn,
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
	var files []string
	if b.config.RemoteType != "" && b.config.Format != "" {
		dir = new(vmwcommon.LocalOutputDir)
		dir.SetOutputDir(b.config.OutputDir)
		files, err = dir.ListFiles()
	} else {
		files, err = state.Get("dir").(OutputDir).ListFiles()
	}
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

	return interpolate.Validate(string(data), &b.config.ctx)
}
