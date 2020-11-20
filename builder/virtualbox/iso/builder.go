//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package iso

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

const BuilderId = "mitchellh.virtualbox"

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig             `mapstructure:",squash"`
	commonsteps.HTTPConfig          `mapstructure:",squash"`
	commonsteps.ISOConfig           `mapstructure:",squash"`
	commonsteps.FloppyConfig        `mapstructure:",squash"`
	commonsteps.CDConfig            `mapstructure:",squash"`
	bootcommand.BootConfig          `mapstructure:",squash"`
	vboxcommon.ExportConfig         `mapstructure:",squash"`
	vboxcommon.OutputConfig         `mapstructure:",squash"`
	vboxcommon.RunConfig            `mapstructure:",squash"`
	vboxcommon.ShutdownConfig       `mapstructure:",squash"`
	vboxcommon.CommConfig           `mapstructure:",squash"`
	vboxcommon.HWConfig             `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig     `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig    `mapstructure:",squash"`
	vboxcommon.VBoxBundleConfig     `mapstructure:",squash"`
	vboxcommon.GuestAdditionsConfig `mapstructure:",squash"`
	// The size, in megabytes, of the hard disk to create for the VM. By
	// default, this is 40000 (about 40 GB).
	DiskSize uint `mapstructure:"disk_size" required:"false"`
	// The guest OS type being installed. By default this is other, but you can
	// get dramatic performance improvements by setting this to the proper
	// value. To view all available values for this run VBoxManage list
	// ostypes. Setting the correct value hints to VirtualBox how to optimize
	// the virtual hardware to work best with that operating system.
	GuestOSType string `mapstructure:"guest_os_type" required:"false"`
	// When this value is set to true, a VDI image will be shrunk in response
	// to the trim command from the guest OS. The size of the cleared area must
	// be at least 1MB. Also set hard_drive_nonrotational to true to enable
	// TRIM support.
	HardDriveDiscard bool `mapstructure:"hard_drive_discard" required:"false"`
	// The type of controller that the primary hard drive is attached to,
	// defaults to ide. When set to sata, the drive is attached to an AHCI SATA
	// controller. When set to scsi, the drive is attached to an LsiLogic SCSI
	// controller. When set to pcie, the drive is attached to an NVMe
	// controller. Please note that when you use "pcie", you'll need to have
	// Virtualbox 6, install an [extension
	// pack](https://www.virtualbox.org/wiki/Downloads#VirtualBox6.0.14OracleVMVirtualBoxExtensionPack)
	// and you will need to enable EFI mode for nvme to work, ex:
	//
	// In JSON:
	// ```json
	//  "vboxmanage": [
	//       [ "modifyvm", "{{.Name}}", "--firmware", "EFI" ],
	//  ]
	// ```
	//
	// In HCL2:
	// ```hcl
	//  vboxmanage = [
	//       [ "modifyvm", "{{.Name}}", "--firmware", "EFI" ],
	//  ]
	// ```
	//
	HardDriveInterface string `mapstructure:"hard_drive_interface" required:"false"`
	// The number of ports available on any SATA controller created, defaults
	// to 1. VirtualBox supports up to 30 ports on a maximum of 1 SATA
	// controller. Increasing this value can be useful if you want to attach
	// additional drives.
	SATAPortCount int `mapstructure:"sata_port_count" required:"false"`
	// The number of ports available on any NVMe controller created, defaults
	// to 1. VirtualBox supports up to 255 ports on a maximum of 1 NVMe
	// controller. Increasing this value can be useful if you want to attach
	// additional drives.
	NVMePortCount int `mapstructure:"nvme_port_count" required:"false"`
	// Forces some guests (i.e. Windows 7+) to treat disks as SSDs and stops
	// them from performing disk fragmentation. Also set hard_drive_discard to
	// true to enable TRIM support.
	HardDriveNonrotational bool `mapstructure:"hard_drive_nonrotational" required:"false"`
	// The type of controller that the ISO is attached to, defaults to ide.
	// When set to sata, the drive is attached to an AHCI SATA controller.
	ISOInterface string `mapstructure:"iso_interface" required:"false"`
	// Set this to true if you would like to keep the VM registered with
	// virtualbox. Defaults to false.
	KeepRegistered bool `mapstructure:"keep_registered" required:"false"`
	// Defaults to false. When enabled, Packer will not export the VM. Useful
	// if the build output is not the resultant image, but created inside the
	// VM.
	SkipExport bool `mapstructure:"skip_export" required:"false"`
	// This is the name of the OVF file for the new virtual machine, without
	// the file extension. By default this is packer-BUILDNAME, where
	// "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name" required:"false"`

	ctx interpolate.Context
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         vboxcommon.BuilderId, // "mitchellh.virtualbox"
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"guest_additions_path",
				"guest_additions_url",
				"vboxmanage",
				"vboxmanage_post",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	isoWarnings, isoErrs := b.config.ISOConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)

	errs = packer.MultiErrorAppend(errs, b.config.ExportConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ExportConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.FloppyConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.CDConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(
		errs, b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.CommConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.HWConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VBoxBundleConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VBoxManageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.VBoxVersionConfig.Prepare(b.config.CommConfig.Comm.Type)...)
	errs = packer.MultiErrorAppend(errs, b.config.BootConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.GuestAdditionsConfig.Prepare(b.config.CommConfig.Comm.Type)...)

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.HardDriveInterface == "" {
		b.config.HardDriveInterface = "ide"
	}

	if b.config.GuestOSType == "" {
		b.config.GuestOSType = "Other"
	}

	if b.config.ISOInterface == "" {
		b.config.ISOInterface = "ide"
	}

	if b.config.GuestAdditionsInterface == "" {
		b.config.GuestAdditionsInterface = b.config.ISOInterface
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf(
			"packer-%s-%d", b.config.PackerBuildName, interpolate.InitTime.Unix())
	}

	switch b.config.HardDriveInterface {
	case "ide", "sata", "scsi", "pcie":
		// do nothing
	default:
		errs = packer.MultiErrorAppend(
			errs, errors.New("hard_drive_interface can only be ide, sata, pcie or scsi"))
	}

	if b.config.SATAPortCount == 0 {
		b.config.SATAPortCount = 1
	}

	if b.config.SATAPortCount > 30 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("sata_port_count cannot be greater than 30"))
	}

	if b.config.NVMePortCount == 0 {
		b.config.NVMePortCount = 1
	}

	if b.config.NVMePortCount > 255 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("nvme_port_count cannot be greater than 255"))
	}

	if b.config.ISOInterface != "ide" && b.config.ISOInterface != "sata" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("iso_interface can only be ide or sata"))
	}

	// Warnings
	if b.config.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with VirtualBox
	driver, err := vboxcommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating VirtualBox driver: %s", err)
	}

	steps := []multistep.Step{
		&vboxcommon.StepDownloadGuestAdditions{
			GuestAdditionsMode:   b.config.GuestAdditionsMode,
			GuestAdditionsURL:    b.config.GuestAdditionsURL,
			GuestAdditionsSHA256: b.config.GuestAdditionsSHA256,
			Ctx:                  b.config.ctx,
		},
		&commonsteps.StepDownload{
			Checksum:    b.config.ISOChecksum,
			Description: "ISO",
			Extension:   b.config.TargetExtension,
			ResultKey:   "iso_path",
			TargetPath:  b.config.TargetPath,
			Url:         b.config.ISOUrls,
		},
		&commonsteps.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&commonsteps.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
			Label:       b.config.FloppyConfig.FloppyLabel,
		},
		&commonsteps.StepCreateCD{
			Files: b.config.CDConfig.CDFiles,
			Label: b.config.CDConfig.CDLabel,
		},
		new(vboxcommon.StepHTTPIPDiscover),
		&commonsteps.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
			HTTPAddress: b.config.HTTPAddress,
		},
		&vboxcommon.StepSshKeyPair{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("%s.pem", b.config.PackerBuildName),
			Comm:         &b.config.Comm,
		},
		new(vboxcommon.StepSuppressMessages),
		new(stepCreateVM),
		new(stepCreateDisk),
		&vboxcommon.StepAttachISOs{
			AttachBootISO:           true,
			ISOInterface:            b.config.ISOInterface,
			GuestAdditionsMode:      b.config.GuestAdditionsMode,
			GuestAdditionsInterface: b.config.GuestAdditionsInterface,
		},
		&vboxcommon.StepConfigureVRDP{
			VRDPBindAddress: b.config.VRDPBindAddress,
			VRDPPortMin:     b.config.VRDPPortMin,
			VRDPPortMax:     b.config.VRDPPortMax,
		},
		new(vboxcommon.StepAttachFloppy),
		&vboxcommon.StepPortForwarding{
			CommConfig:     &b.config.CommConfig.Comm,
			HostPortMin:    b.config.HostPortMin,
			HostPortMax:    b.config.HostPortMax,
			SkipNatMapping: b.config.SkipNatMapping,
		},
		&vboxcommon.StepVBoxManage{
			Commands: b.config.VBoxManage,
			Ctx:      b.config.ctx,
		},
		&vboxcommon.StepRun{
			Headless: b.config.Headless,
		},
		&vboxcommon.StepTypeBootCommand{
			BootWait:      b.config.BootWait,
			BootCommand:   b.config.FlatBootCommand(),
			VMName:        b.config.VMName,
			Ctx:           b.config.ctx,
			GroupInterval: b.config.BootConfig.BootGroupInterval,
			Comm:          &b.config.Comm,
		},
		&communicator.StepConnect{
			Config:    &b.config.CommConfig.Comm,
			Host:      vboxcommon.CommHost(b.config.CommConfig.Comm.SSHHost),
			SSHConfig: b.config.CommConfig.Comm.SSHConfigFunc(),
			SSHPort:   vboxcommon.CommPort,
			WinRMPort: vboxcommon.CommPort,
		},
		&vboxcommon.StepUploadVersion{
			Path: *b.config.VBoxVersionFile,
		},
		&vboxcommon.StepUploadGuestAdditions{
			GuestAdditionsMode: b.config.GuestAdditionsMode,
			GuestAdditionsPath: b.config.GuestAdditionsPath,
			Ctx:                b.config.ctx,
		},
		new(commonsteps.StepProvision),
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.CommConfig.Comm,
		},
		&vboxcommon.StepShutdown{
			Command:         b.config.ShutdownCommand,
			Timeout:         b.config.ShutdownTimeout,
			Delay:           b.config.PostShutdownDelay,
			DisableShutdown: b.config.DisableShutdown,
			ACPIShutdown:    b.config.ACPIShutdown,
		},
		&vboxcommon.StepRemoveDevices{
			Bundling: b.config.VBoxBundleConfig,
		},
		&vboxcommon.StepVBoxManage{
			Commands: b.config.VBoxManagePost,
			Ctx:      b.config.ctx,
		},
		&vboxcommon.StepExport{
			Format:         b.config.Format,
			OutputDir:      b.config.OutputDir,
			OutputFilename: b.config.OutputFilename,
			ExportOpts:     b.config.ExportConfig.ExportOpts,
			Bundling:       b.config.VBoxBundleConfig,
			SkipNatMapping: b.config.SkipNatMapping,
			SkipExport:     b.config.SkipExport,
		},
	}

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

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

	generatedData := map[string]interface{}{"generated_data": state.Get("generated_data")}
	return vboxcommon.NewArtifact(b.config.OutputDir, generatedData)
}
