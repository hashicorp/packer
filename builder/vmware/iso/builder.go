package iso

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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
	common.FloppyConfig      `mapstructure:",squash"`
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
	Format              string   `mapstructure:"format"`
	GuestOSType         string   `mapstructure:"guest_os_type"`
	KeepRegistered      bool     `mapstructure:"keep_registered"`
	OVFToolOptions      []string `mapstructure:"ovftool_options"`
	SkipCompaction      bool     `mapstructure:"skip_compaction"`
	SkipExport          bool     `mapstructure:"skip_export"`
	VMName              string   `mapstructure:"vm_name"`
	VMXDiskTemplatePath string   `mapstructure:"vmx_disk_template_path"`
	VMXTemplatePath     string   `mapstructure:"vmx_template_path"`
	Version             string   `mapstructure:"version"`

	RemoteType           string `mapstructure:"remote_type"`
	RemoteDatastore      string `mapstructure:"remote_datastore"`
	RemoteCacheDatastore string `mapstructure:"remote_cache_datastore"`
	RemoteCacheDirectory string `mapstructure:"remote_cache_directory"`
	RemoteHost           string `mapstructure:"remote_host"`
	RemotePort           uint   `mapstructure:"remote_port"`
	RemoteUser           string `mapstructure:"remote_username"`
	RemotePassword       string `mapstructure:"remote_password"`
	RemotePrivateKey     string `mapstructure:"remote_private_key_file"`

	CommConfig communicator.Config `mapstructure:",squash"`

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
	errs = packer.MultiErrorAppend(errs, b.config.FloppyConfig.Prepare(&b.config.ctx)...)

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

	if b.config.Headless && b.config.DisableVNC {
		warnings = append(warnings,
			"Headless mode uses VNC to retrieve output. Since VNC has been disabled,\n"+
				"you won't be able to see any output.")
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

	exportOutputPath := b.config.OutputDir

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
			Extension:    b.config.TargetExtension,
			ResultKey:    "iso_path",
			TargetPath:   b.config.TargetPath,
			Url:          b.config.ISOUrls,
		},
		&vmwcommon.StepOutputDir{
			Force: b.config.PackerForce,
		},
		&common.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
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
			Enabled:            !b.config.DisableVNC,
			VNCBindAddress:     b.config.VNCBindAddress,
			VNCPortMin:         b.config.VNCPortMin,
			VNCPortMax:         b.config.VNCPortMax,
			VNCDisablePassword: b.config.VNCDisablePassword,
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
			VNCEnabled:  !b.config.DisableVNC,
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
		&vmwcommon.StepCleanVMX{
			RemoveEthernetInterfaces: b.config.VMXConfig.VMXRemoveEthernet,
			VNCEnabled:               !b.config.DisableVNC,
		},
		&StepUploadVMX{
			RemoteType: b.config.RemoteType,
		},
		&StepExport{
			Format:     b.config.Format,
			SkipExport: b.config.SkipExport,
			OutputDir:  exportOutputPath,
		},
	}

	// Run!
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
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
	if b.config.RemoteType != "" && b.config.Format != "" && !b.config.SkipExport {
		dir = new(vmwcommon.LocalOutputDir)
		dir.SetOutputDir(exportOutputPath)
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

	config := make(map[string]string)
	config[ArtifactConfKeepRegistered] = strconv.FormatBool(b.config.KeepRegistered)
	config[ArtifactConfFormat] = b.config.Format
	config[ArtifactConfSkipExport] = strconv.FormatBool(b.config.SkipExport)

	return &Artifact{
		builderId: builderId,
		id:        b.config.VMName,
		dir:       dir,
		f:         files,
		config:    config,
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
