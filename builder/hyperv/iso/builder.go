//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package iso

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	hypervcommon "github.com/hashicorp/packer/builder/hyperv/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/common/shutdowncommand"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const (
	DefaultDiskSize = 40 * 1024        // ~40GB
	MinDiskSize     = 256              // 256MB
	MaxDiskSize     = 64 * 1024 * 1024 // 64TB
	MaxVHDSize      = 2040 * 1024      // 2040GB

	DefaultDiskBlockSize = 32  // 32MB
	MinDiskBlockSize     = 1   // 1MB
	MaxDiskBlockSize     = 256 // 256MB

	DefaultRamSize                 = 1 * 1024  // 1GB
	MinRamSize                     = 32        // 32MB
	MaxRamSize                     = 32 * 1024 // 32GB
	MinNestedVirtualizationRamSize = 4 * 1024  // 4GB

	LowRam = 256 // 256MB

	DefaultUsername = ""
	DefaultPassword = ""
)

// Builder implements packer.Builder and builds the actual Hyperv
// images.
type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig            `mapstructure:",squash"`
	common.HTTPConfig              `mapstructure:",squash"`
	common.ISOConfig               `mapstructure:",squash"`
	bootcommand.BootConfig         `mapstructure:",squash"`
	hypervcommon.OutputConfig      `mapstructure:",squash"`
	hypervcommon.SSHConfig         `mapstructure:",squash"`
	hypervcommon.CommonConfig      `mapstructure:",squash"`
	shutdowncommand.ShutdownConfig `mapstructure:",squash"`
	// The size, in megabytes, of the hard disk to create
	// for the VM. By default, this is 40 GB.
	DiskSize uint `mapstructure:"disk_size" required:"false"`
	// If true use a legacy network adapter as the NIC.
	// This defaults to false. A legacy network adapter is fully emulated NIC, and is thus
	// supported by various exotic operating systems, but this emulation requires
	// additional overhead and should only be used if absolutely necessary.
	UseLegacyNetworkAdapter bool `mapstructure:"use_legacy_network_adapter" required:"false"`
	// If true enables differencing disks. Only
	// the changes will be written to the new disk. This is especially useful if
	// your source is a VHD/VHDX. This defaults to false.
	DifferencingDisk bool `mapstructure:"differencing_disk" required:"false"`
	// If true, creates the boot disk on the
	// virtual machine as a fixed VHD format disk. The default is false, which
	// creates a dynamic VHDX format disk. This option requires setting
	// generation to 1, skip_compaction to true, and
	// differencing_disk to false. Additionally, any value entered for
	// disk_block_size will be ignored. The most likely use case for this
	// option is outputing a disk that is in the format required for upload to
	// Azure.
	FixedVHD bool `mapstructure:"use_fixed_vhd_format" required:"false"`

	ctx interpolate.Context
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
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

	errs = packer.MultiErrorAppend(errs, b.config.BootConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)

	commonErrs, commonWarns := b.config.CommonConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)
	packer.MultiErrorAppend(errs, commonErrs...)
	warnings = append(warnings, commonWarns...)

	if len(b.config.ISOConfig.ISOUrls) < 1 ||
		(strings.ToLower(filepath.Ext(b.config.ISOConfig.ISOUrls[0])) != ".vhd" &&
			strings.ToLower(filepath.Ext(b.config.ISOConfig.ISOUrls[0])) != ".vhdx") {
		//We only create a new hard drive if an existing one to copy from does not exist
		err = b.checkDiskSize()
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if b.config.Cpu < 1 {
		b.config.Cpu = 1
	}

	if b.config.Generation == 2 {
		if b.config.UseLegacyNetworkAdapter {
			err = errors.New("Generation 2 vms don't support legacy network adapters.")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	// Errors

	if b.config.Generation > 1 && b.config.FixedVHD {
		err = errors.New("Fixed VHD disks are only supported on Generation 1 virtual machines.")
		errs = packer.MultiErrorAppend(errs, err)
	}

	if !b.config.SkipCompaction && b.config.FixedVHD {
		err = errors.New("Fixed VHD disks do not support compaction.")
		errs = packer.MultiErrorAppend(errs, err)
	}

	if b.config.DifferencingDisk && b.config.FixedVHD {
		err = errors.New("Fixed VHD disks are not supported with differencing disks.")
		errs = packer.MultiErrorAppend(errs, err)
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

// Run executes a Packer build and returns a packer.Artifact representing
// a Hyperv appliance.
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Hyperv
	driver, err := hypervcommon.NewHypervPS4Driver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Hyper-V driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&hypervcommon.StepCreateBuildDir{
			TempPath: b.config.TempPath,
		},
		&common.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			ResultKey:    "iso_path",
			Url:          b.config.ISOUrls,
			Extension:    b.config.TargetExtension,
			TargetPath:   b.config.TargetPath,
		},
		&common.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
			Label:       b.config.FloppyConfig.FloppyLabel,
		},
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		&hypervcommon.StepCreateSwitch{
			SwitchName: b.config.SwitchName,
		},
		&hypervcommon.StepCreateVM{
			VMName:                         b.config.VMName,
			SwitchName:                     b.config.SwitchName,
			RamSize:                        b.config.RamSize,
			DiskSize:                       b.config.DiskSize,
			DiskBlockSize:                  b.config.DiskBlockSize,
			Generation:                     b.config.Generation,
			Cpu:                            b.config.Cpu,
			EnableMacSpoofing:              b.config.EnableMacSpoofing,
			EnableDynamicMemory:            b.config.EnableDynamicMemory,
			EnableSecureBoot:               b.config.EnableSecureBoot,
			SecureBootTemplate:             b.config.SecureBootTemplate,
			EnableVirtualizationExtensions: b.config.EnableVirtualizationExtensions,
			UseLegacyNetworkAdapter:        b.config.UseLegacyNetworkAdapter,
			AdditionalDiskSize:             b.config.AdditionalDiskSize,
			DifferencingDisk:               b.config.DifferencingDisk,
			MacAddress:                     b.config.MacAddress,
			FixedVHD:                       b.config.FixedVHD,
			Version:                        b.config.Version,
			KeepRegistered:                 b.config.KeepRegistered,
		},
		&hypervcommon.StepEnableIntegrationService{},

		&hypervcommon.StepMountDvdDrive{
			Generation: b.config.Generation,
		},
		&hypervcommon.StepMountFloppydrive{
			Generation: b.config.Generation,
		},

		&hypervcommon.StepMountGuestAdditions{
			GuestAdditionsMode: b.config.GuestAdditionsMode,
			GuestAdditionsPath: b.config.GuestAdditionsPath,
			Generation:         b.config.Generation,
		},

		&hypervcommon.StepMountSecondaryDvdImages{
			IsoPaths:   b.config.SecondaryDvdImages,
			Generation: b.config.Generation,
		},

		&hypervcommon.StepConfigureVlan{
			VlanId:       b.config.VlanId,
			SwitchVlanId: b.config.SwitchVlanId,
		},

		&hypervcommon.StepRun{
			Headless:   b.config.Headless,
			SwitchName: b.config.SwitchName,
		},

		&hypervcommon.StepTypeBootCommand{
			BootCommand:   b.config.FlatBootCommand(),
			BootWait:      b.config.BootWait,
			SwitchName:    b.config.SwitchName,
			Ctx:           b.config.ctx,
			GroupInterval: b.config.BootConfig.BootGroupInterval,
		},

		// configure the communicator ssh, winrm
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      hypervcommon.CommHost(b.config.SSHConfig.Comm.SSHHost),
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
		},

		// provision requires communicator to be setup
		&common.StepProvision{},

		// Remove ephemeral key from authorized_hosts if using SSH communicator
		&common.StepCleanupTempKeys{
			Comm: &b.config.SSHConfig.Comm,
		},

		&hypervcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},

		// wait for the vm to be powered off
		&hypervcommon.StepWaitForPowerOff{},

		// remove the secondary dvd images
		// after we power down
		&hypervcommon.StepUnmountSecondaryDvdImages{},
		&hypervcommon.StepUnmountGuestAdditions{},
		&hypervcommon.StepUnmountDvdDrive{},
		&hypervcommon.StepUnmountFloppyDrive{
			Generation: b.config.Generation,
		},
		&hypervcommon.StepCompactDisk{
			SkipCompaction: b.config.SkipCompaction,
		},
		&hypervcommon.StepExportVm{
			OutputDir:  b.config.OutputDir,
			SkipExport: b.config.SkipExport,
		},
		&hypervcommon.StepCollateArtifacts{
			OutputDir:  b.config.OutputDir,
			SkipExport: b.config.SkipExport,
		},

		// the clean up actions for each step will be executed reverse order
	}

	// Run the steps.
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// Report any errors.
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

	return hypervcommon.NewArtifact(b.config.OutputDir)
}

// Cancel.

func (b *Builder) checkDiskSize() error {
	if b.config.DiskSize == 0 {
		b.config.DiskSize = DefaultDiskSize
	}

	log.Println(fmt.Sprintf("%s: %v", "DiskSize", b.config.DiskSize))

	if b.config.DiskSize < MinDiskSize {
		return fmt.Errorf("disk_size: Virtual machine requires disk space >= %v GB, but defined: %v",
			MinDiskSize, b.config.DiskSize/1024)
	} else if b.config.DiskSize > MaxDiskSize && !b.config.FixedVHD {
		return fmt.Errorf("disk_size: Virtual machine requires disk space <= %v GB, but defined: %v",
			MaxDiskSize, b.config.DiskSize/1024)
	} else if b.config.DiskSize > MaxVHDSize && b.config.FixedVHD {
		return fmt.Errorf("disk_size: Virtual machine requires disk space <= %v GB, but defined: %v",
			MaxVHDSize/1024, b.config.DiskSize/1024)
	}

	return nil
}
