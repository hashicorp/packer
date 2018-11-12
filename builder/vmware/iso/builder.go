package iso

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	common.HTTPConfig        `mapstructure:",squash"`
	common.ISOConfig         `mapstructure:",squash"`
	common.FloppyConfig      `mapstructure:",squash"`
	bootcommand.VNCConfig    `mapstructure:",squash"`
	vmwcommon.DriverConfig   `mapstructure:",squash"`
	vmwcommon.OutputConfig   `mapstructure:",squash"`
	vmwcommon.RunConfig      `mapstructure:",squash"`
	vmwcommon.ShutdownConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig      `mapstructure:",squash"`
	vmwcommon.ToolsConfig    `mapstructure:",squash"`
	vmwcommon.VMXConfig      `mapstructure:",squash"`
	vmwcommon.ExportConfig   `mapstructure:",squash"`

	// disk drives
	AdditionalDiskSize []uint `mapstructure:"disk_additional_size"`
	DiskAdapterType    string `mapstructure:"disk_adapter_type"`
	DiskName           string `mapstructure:"vmdk_name"`
	DiskSize           uint   `mapstructure:"disk_size"`
	DiskTypeId         string `mapstructure:"disk_type_id"`
	Format             string `mapstructure:"format"`

	// cdrom drive
	CdromAdapterType string `mapstructure:"cdrom_adapter_type"`

	// platform information
	GuestOSType string `mapstructure:"guest_os_type"`
	Version     string `mapstructure:"version"`
	VMName      string `mapstructure:"vm_name"`

	// Network adapter and type
	NetworkAdapterType string `mapstructure:"network_adapter_type"`
	Network            string `mapstructure:"network"`

	// device presence
	Sound bool `mapstructure:"sound"`
	USB   bool `mapstructure:"usb"`

	// communication ports
	Serial   string `mapstructure:"serial"`
	Parallel string `mapstructure:"parallel"`

	VMXDiskTemplatePath string `mapstructure:"vmx_disk_template_path"`
	VMXTemplatePath     string `mapstructure:"vmx_template_path"`

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
	errs = packer.MultiErrorAppend(errs, b.config.VNCConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ExportConfig.Prepare(&b.config.ctx)...)

	if b.config.DiskName == "" {
		b.config.DiskName = "disk"
	}

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.DiskAdapterType == "" {
		// Default is lsilogic
		b.config.DiskAdapterType = "lsilogic"
	}

	if !b.config.SkipCompaction {
		if b.config.RemoteType == "esx5" {
			if b.config.DiskTypeId == "" {
				b.config.SkipCompaction = true
			}
		}
	}

	if b.config.DiskTypeId == "" {
		// Default is growable virtual disk split in 2GB files.
		b.config.DiskTypeId = "1"

		if b.config.RemoteType == "esx5" {
			b.config.DiskTypeId = "zeroedthick"
		}
	}

	if b.config.RemoteType == "esx5" {
		if b.config.DiskTypeId != "thin" && !b.config.SkipCompaction {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("skip_compaction must be 'true' for disk_type_id: %s", b.config.DiskTypeId))
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

	if b.config.VMXTemplatePath != "" {
		if err := b.validateVMXTemplatePath(); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("vmx_template_path is invalid: %s", err))
		}

	} else {
		warn := b.checkForVMXTemplateAndVMXDataCollisions()
		if warn != "" {
			warnings = append(warnings, warn)
		}
	}

	if b.config.Network == "" {
		b.config.Network = "nat"
	}

	if !b.config.Sound {
		b.config.Sound = false
	}

	if !b.config.USB {
		b.config.USB = false
	}

	// Remote configuration validation
	if b.config.RemoteType != "" {
		if b.config.RemoteHost == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("remote_host must be specified"))
		}

		if b.config.RemoteType != "esx5" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Only 'esx5' value is accepted for remote_type"))
		}
	}

	if b.config.Format == "" {
		b.config.Format = "ovf"
	}

	if !(b.config.Format == "ova" || b.config.Format == "ovf" || b.config.Format == "vmx") {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("format must be one of ova, ovf, or vmx"))
	}

	if b.config.RemoteType == "esx5" && b.config.SkipExport != true && b.config.RemotePassword == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("exporting the vm (with ovftool) requires that you set a value for remote_password"))
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
	driver, err := vmwcommon.NewDriver(&b.config.DriverConfig, &b.config.SSHConfig, b.config.VMName)
	if err != nil {
		return nil, fmt.Errorf("Failed creating VMware driver: %s", err)
	}

	// Determine the output dir implementation
	var dir vmwcommon.OutputDir
	switch d := driver.(type) {
	case vmwcommon.OutputDir:
		dir = d
	default:
		dir = new(vmwcommon.LocalOutputDir)
	}

	// The OutputDir will track remote esxi output; exportOutputPath preserves
	// the path to the output on the machine running Packer.
	exportOutputPath := b.config.OutputDir

	if b.config.RemoteType != "" {
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
	state.Put("sshConfig", &b.config.SSHConfig)
	state.Put("driverConfig", &b.config.DriverConfig)

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
			Key:       "floppy_path",
			Message:   "Uploading Floppy to remote machine...",
			DoCleanup: true,
		},
		&stepRemoteUpload{
			Key:     "iso_path",
			Message: "Uploading ISO to remote machine...",
		},
		&stepCreateDisk{},
		&stepCreateVMX{},
		&vmwcommon.StepConfigureVMX{
			CustomData:  b.config.VMXData,
			VMName:      b.config.VMName,
			DisplayName: b.config.VMXDisplayName,
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
		&vmwcommon.StepRegister{
			Format:         b.config.Format,
			KeepRegistered: b.config.KeepRegistered,
			SkipExport:     b.config.SkipExport,
		},
		&vmwcommon.StepRun{
			DurationBeforeStop: 5 * time.Second,
			Headless:           b.config.Headless,
		},
		&vmwcommon.StepTypeBootCommand{
			BootWait:    b.config.BootWait,
			VNCEnabled:  !b.config.DisableVNC,
			BootCommand: b.config.FlatBootCommand(),
			VMName:      b.config.VMName,
			Ctx:         b.config.ctx,
			KeyInterval: b.config.VNCConfig.BootKeyInterval,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      driver.CommHost,
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
		},
		&vmwcommon.StepUploadTools{
			RemoteType:        b.config.RemoteType,
			ToolsUploadFlavor: b.config.ToolsUploadFlavor,
			ToolsUploadPath:   b.config.ToolsUploadPath,
			Ctx:               b.config.ctx,
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.SSHConfig.Comm,
		},
		&vmwcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&vmwcommon.StepCleanFiles{},
		&vmwcommon.StepCompactDisk{
			Skip: b.config.SkipCompaction,
		},
		&vmwcommon.StepConfigureVMX{
			CustomData:  b.config.VMXDataPost,
			SkipFloppy:  true,
			VMName:      b.config.VMName,
			DisplayName: b.config.VMXDisplayName,
		},
		&vmwcommon.StepCleanVMX{
			RemoveEthernetInterfaces: b.config.VMXConfig.VMXRemoveEthernet,
			VNCEnabled:               !b.config.DisableVNC,
		},
		&vmwcommon.StepUploadVMX{
			RemoteType: b.config.RemoteType,
		},
		&vmwcommon.StepExport{
			Format:         b.config.Format,
			SkipExport:     b.config.SkipExport,
			VMName:         b.config.VMName,
			OVFToolOptions: b.config.OVFToolOptions,
			OutputDir:      exportOutputPath,
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
	return vmwcommon.NewArtifact(b.config.RemoteType, b.config.Format, exportOutputPath,
		b.config.VMName, b.config.SkipExport, b.config.KeepRegistered, state)
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

// Validate the vmx_data option against the default vmx template to warn
// user if anything is being overridden.
func (b *Builder) checkForVMXTemplateAndVMXDataCollisions() string {
	if b.config.VMXTemplatePath != "" {
		return ""
	}

	var overridden []string
	tplLines := strings.Split(DefaultVMXTemplate, "\n")
	tplLines = append(tplLines,
		fmt.Sprintf("%s0:0.present", strings.ToLower(b.config.DiskAdapterType)),
		fmt.Sprintf("%s0:0.fileName", strings.ToLower(b.config.DiskAdapterType)),
		fmt.Sprintf("%s0:0.deviceType", strings.ToLower(b.config.DiskAdapterType)),
		fmt.Sprintf("%s0:1.present", strings.ToLower(b.config.DiskAdapterType)),
		fmt.Sprintf("%s0:1.fileName", strings.ToLower(b.config.DiskAdapterType)),
		fmt.Sprintf("%s0:1.deviceType", strings.ToLower(b.config.DiskAdapterType)),
	)

	for _, line := range tplLines {
		if strings.Contains(line, `{{`) {
			key := line[:strings.Index(line, " =")]
			if _, ok := b.config.VMXData[key]; ok {
				overridden = append(overridden, key)
			}
		}
	}

	if len(overridden) > 0 {
		warnings := fmt.Sprintf("Your vmx data contains the following "+
			"variable(s), which Packer normally sets when it generates its "+
			"own default vmx template. This may cause your build to fail or "+
			"behave unpredictably: %s", strings.Join(overridden, ", "))
		return warnings
	}
	return ""
}

// Make sure custom vmx template exists and that data can be read from it
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
