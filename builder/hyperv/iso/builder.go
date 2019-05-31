//go:generate struct-markdown

package iso

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	hypervcommon "github.com/hashicorp/packer/builder/hyperv/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	powershell "github.com/hashicorp/packer/common/powershell"
	"github.com/hashicorp/packer/common/powershell/hyperv"
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
	common.PackerConfig         `mapstructure:",squash"`
	common.HTTPConfig           `mapstructure:",squash"`
	common.ISOConfig            `mapstructure:",squash"`
	common.FloppyConfig         `mapstructure:",squash"`
	bootcommand.BootConfig      `mapstructure:",squash"`
	hypervcommon.OutputConfig   `mapstructure:",squash"`
	hypervcommon.SSHConfig      `mapstructure:",squash"`
	hypervcommon.ShutdownConfig `mapstructure:",squash"`
	// The size, in megabytes, of the hard disk to create
    // for the VM. By default, this is 40 GB.
	DiskSize uint `mapstructure:"disk_size" required:"false"`
	// The block size of the VHD to be created.
    // Recommended disk block size for Linux hyper-v guests is 1 MiB. This
    // defaults to "32 MiB".
	DiskBlockSize uint `mapstructure:"disk_block_size" required:"false"`
	// The amount, in megabytes, of RAM to assign to the
    // VM. By default, this is 1 GB.
	RamSize uint `mapstructure:"memory" required:"false"`
	// A list of ISO paths to
    // attach to a VM when it is booted. This is most useful for unattended
    // Windows installs, which look for an Autounattend.xml file on removable
    // media. By default, no secondary ISO will be attached.
	SecondaryDvdImages []string `mapstructure:"secondary_iso_images" required:"false"`
	// If set to attach then attach and
    // mount the ISO image specified in guest_additions_path. If set to
    // none then guest additions are not attached and mounted; This is the
    // default.
	GuestAdditionsMode string `mapstructure:"guest_additions_mode" required:"false"`
	// The path to the ISO image for guest
    // additions.
	GuestAdditionsPath string `mapstructure:"guest_additions_path" required:"false"`
	// This is the name of the new virtual machine,
    // without the file extension. By default this is "packer-BUILDNAME",
    // where "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name" required:"false"`
	// The name of the switch to connect the virtual
    // machine to. By default, leaving this value unset will cause Packer to
    // try and determine the switch to use by looking for an external switch
    // that is up and running.
	SwitchName                     string `mapstructure:"switch_name" required:"false"`
	// This is the VLAN of the virtual switch's
    // network card. By default none is set. If none is set then a VLAN is not
    // set on the switch's network card. If this value is set it should match
    // the VLAN specified in by vlan_id.
	SwitchVlanId                   string `mapstructure:"switch_vlan_id" required:"false"`
	// This allows a specific MAC address to be used on
    // the default virtual network card. The MAC address must be a string with
    // no delimiters, for example "0000deadbeef".
	MacAddress                     string `mapstructure:"mac_address" required:"false"`
	// This is the VLAN of the virtual machine's network
    // card for the new virtual machine. By default none is set. If none is set
    // then VLANs are not set on the virtual machine's network card.
	VlanId                         string `mapstructure:"vlan_id" required:"false"`
	// The number of CPUs the virtual machine should use. If
    // this isn't specified, the default is 1 CPU.
	Cpu                            uint   `mapstructure:"cpus" required:"false"`
	// The Hyper-V generation for the virtual machine. By
    // default, this is 1. Generation 2 Hyper-V virtual machines do not support
    // floppy drives. In this scenario use secondary_iso_images instead. Hard
    // drives and DVD drives will also be SCSI and not IDE.
	Generation                     uint   `mapstructure:"generation" required:"false"`
	// If true enable MAC address spoofing
    // for the virtual machine. This defaults to false.
	EnableMacSpoofing              bool   `mapstructure:"enable_mac_spoofing" required:"false"`
	// If true use a legacy network adapter as the NIC.
    // This defaults to false. A legacy network adapter is fully emulated NIC, and is thus
    // supported by various exotic operating systems, but this emulation requires
    // additional overhead and should only be used if absolutely necessary.
	UseLegacyNetworkAdapter        bool   `mapstructure:"use_legacy_network_adapter" required:"false"`
	// If true enable dynamic memory for
    // the virtual machine. This defaults to false.
	EnableDynamicMemory            bool   `mapstructure:"enable_dynamic_memory" required:"false"`
	// If true enable secure boot for the
    // virtual machine. This defaults to false. See secure_boot_template
    // below for additional settings.
	EnableSecureBoot               bool   `mapstructure:"enable_secure_boot" required:"false"`
	// The secure boot template to be
    // configured. Valid values are "MicrosoftWindows" (Windows) or
    // "MicrosoftUEFICertificateAuthority" (Linux). This only takes effect if
    // enable_secure_boot is set to "true". This defaults to "MicrosoftWindows".
	SecureBootTemplate             string `mapstructure:"secure_boot_template" required:"false"`
	// If true enable
    // virtualization extensions for the virtual machine. This defaults to
    // false. For nested virtualization you need to enable MAC spoofing,
    // disable dynamic memory and have at least 4GB of RAM assigned to the
    // virtual machine.
	EnableVirtualizationExtensions bool   `mapstructure:"enable_virtualization_extensions" required:"false"`
	// The location under which Packer will create a
    // directory to house all the VM files and folders during the build.
    // By default %TEMP% is used which, for most systems, will evaluate to
    // %USERPROFILE%/AppData/Local/Temp.
	TempPath                       string `mapstructure:"temp_path" required:"false"`
	// This allows you to set the vm version when
    //  calling New-VM to generate the vm.
	Version                        string `mapstructure:"configuration_version" required:"false"`
	// If "true", Packer will not delete the VM from
    // The Hyper-V manager.
	KeepRegistered                 bool   `mapstructure:"keep_registered" required:"false"`

	Communicator string `mapstructure:"communicator"`
	// The size or sizes of any
    // additional hard disks for the VM in megabytes. If this is not specified
    // then the VM will only contain a primary hard disk. Additional drives
    // will be attached to the SCSI interface only. The builder uses
    // expandable rather than fixed-size virtual hard disks, so the actual
    // file representing the disk will not use the full size unless it is
    // full.
	AdditionalDiskSize []uint `mapstructure:"disk_additional_size" required:"false"`
	// If true skip compacting the hard disk for
    // the virtual machine when exporting. This defaults to false.
	SkipCompaction bool `mapstructure:"skip_compaction" required:"false"`
	// If true Packer will skip the export of the VM.
    // If you are interested only in the VHD/VHDX files, you can enable this
    // option. The resulting VHD/VHDX file will be output to
    // <output_directory>/Virtual Hard Disks. By default this option is false
    // and Packer will export the VM to output_directory.
	SkipExport bool `mapstructure:"skip_export" required:"false"`
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
	// Packer defaults to building Hyper-V virtual
    // machines by launching a GUI that shows the console of the machine being
    // built. When this value is set to true, the machine will start without a
    // console.
	Headless bool `mapstructure:"headless" required:"false"`

	ctx interpolate.Context
}

// Prepare processes the build configuration parameters.
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
	errs = packer.MultiErrorAppend(errs, b.config.FloppyConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.OutputConfig.Prepare(&b.config.ctx, &b.config.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, b.config.SSHConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)

	if len(b.config.ISOConfig.ISOUrls) < 1 ||
		(strings.ToLower(filepath.Ext(b.config.ISOConfig.ISOUrls[0])) != ".vhd" &&
			strings.ToLower(filepath.Ext(b.config.ISOConfig.ISOUrls[0])) != ".vhdx") {
		//We only create a new hard drive if an existing one to copy from does not exist
		err = b.checkDiskSize()
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	err = b.checkDiskBlockSize()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	err = b.checkRamSize()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}

	log.Println(fmt.Sprintf("%s: %v", "VMName", b.config.VMName))

	if b.config.SwitchName == "" {
		b.config.SwitchName = b.detectSwitchName()
	}

	if b.config.Cpu < 1 {
		b.config.Cpu = 1
	}

	if b.config.Generation < 1 || b.config.Generation > 2 {
		b.config.Generation = 1
	}

	if b.config.Generation == 2 {
		if len(b.config.FloppyFiles) > 0 || len(b.config.FloppyDirectories) > 0 {
			err = errors.New("Generation 2 vms don't support floppy drives. Use ISO image instead.")
			errs = packer.MultiErrorAppend(errs, err)
		}
		if b.config.UseLegacyNetworkAdapter {
			err = errors.New("Generation 2 vms don't support legacy network adapters.")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if len(b.config.AdditionalDiskSize) > 64 {
		err = errors.New("VM's currently support a maximum of 64 additional SCSI attached disks.")
		errs = packer.MultiErrorAppend(errs, err)
	}

	log.Println(fmt.Sprintf("Using switch %s", b.config.SwitchName))
	log.Println(fmt.Sprintf("%s: %v", "SwitchName", b.config.SwitchName))

	// Errors

	if b.config.GuestAdditionsMode == "" {
		if b.config.GuestAdditionsPath != "" {
			b.config.GuestAdditionsMode = "attach"
		} else {
			b.config.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

			if _, err := os.Stat(b.config.GuestAdditionsPath); os.IsNotExist(err) {
				if err != nil {
					b.config.GuestAdditionsPath = ""
					b.config.GuestAdditionsMode = "none"
				} else {
					b.config.GuestAdditionsMode = "attach"
				}
			}
		}
	}

	if b.config.GuestAdditionsPath == "" && b.config.GuestAdditionsMode == "attach" {
		b.config.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

		if _, err := os.Stat(b.config.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				b.config.GuestAdditionsPath = ""
			}
		}
	}

	for _, isoPath := range b.config.SecondaryDvdImages {
		if _, err := os.Stat(isoPath); os.IsNotExist(err) {
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Secondary Dvd image does not exist: %s", err))
			}
		}
	}

	numberOfIsos := len(b.config.SecondaryDvdImages)

	if b.config.GuestAdditionsMode == "attach" {
		if _, err := os.Stat(b.config.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Guest additions iso does not exist: %s", err))
			}
		}

		numberOfIsos = numberOfIsos + 1
	}

	if b.config.Generation < 2 && numberOfIsos > 2 {
		if b.config.GuestAdditionsMode == "attach" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are only 2 ide controllers available, "+
				"so we can't support guest additions and these secondary dvds: %s",
				strings.Join(b.config.SecondaryDvdImages, ", ")))
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are only 2 ide controllers available, "+
				"so we can't support these secondary dvds: %s", strings.Join(b.config.SecondaryDvdImages, ", ")))
		}
	} else if b.config.Generation > 1 && len(b.config.SecondaryDvdImages) > 16 {
		if b.config.GuestAdditionsMode == "attach" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are not enough drive letters available "+
				"for scsi (limited to 16), so we can't support guest additions and these secondary dvds: %s",
				strings.Join(b.config.SecondaryDvdImages, ", ")))
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("There are not enough drive letters available "+
				"for scsi (limited to 16), so we can't support these secondary dvds: %s",
				strings.Join(b.config.SecondaryDvdImages, ", ")))
		}
	}

	if b.config.EnableVirtualizationExtensions {
		hasVirtualMachineVirtualizationExtensions, err := powershell.HasVirtualMachineVirtualizationExtensions()
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed detecting virtual machine virtualization "+
				"extensions support: %s", err))
		} else {
			if !hasVirtualMachineVirtualizationExtensions {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("This version of Hyper-V does not support "+
					"virtual machine virtualization extension. Please use Windows 10 or Windows Server "+
					"2016 or newer."))
			}
		}
	}

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

	warning := b.checkHostAvailableMemory()
	if warning != "" {
		warnings = appendWarnings(warnings, warning)
	}

	if b.config.EnableVirtualizationExtensions {
		if b.config.EnableDynamicMemory {
			warning = fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, " +
				"dynamic memory should not be allowed.")
			warnings = appendWarnings(warnings, warning)
		}

		if !b.config.EnableMacSpoofing {
			warning = fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, " +
				"mac spoofing should be allowed.")
			warnings = appendWarnings(warnings, warning)
		}

		if b.config.RamSize < MinNestedVirtualizationRamSize {
			warning = fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, " +
				"there should be 4GB or more memory set for the vm, otherwise Hyper-V may fail to start " +
				"any nested VMs.")
			warnings = appendWarnings(warnings, warning)
		}
	}

	if b.config.SwitchVlanId != "" {
		if b.config.SwitchVlanId != b.config.VlanId {
			warning = fmt.Sprintf("Switch network adaptor vlan should match virtual machine network adaptor " +
				"vlan. The switch will not be able to see traffic from the VM.")
			warnings = appendWarnings(warnings, warning)
		}
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
			Headless: b.config.Headless,
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

func appendWarnings(slice []string, data ...string) []string {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]string, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

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

func (b *Builder) checkDiskBlockSize() error {
	if b.config.DiskBlockSize == 0 {
		b.config.DiskBlockSize = DefaultDiskBlockSize
	}

	log.Println(fmt.Sprintf("%s: %v", "DiskBlockSize", b.config.DiskBlockSize))

	if b.config.DiskBlockSize < MinDiskBlockSize {
		return fmt.Errorf("disk_block_size: Virtual machine requires disk block size >= %v MB, but defined: %v",
			MinDiskBlockSize, b.config.DiskBlockSize)
	} else if b.config.DiskBlockSize > MaxDiskBlockSize {
		return fmt.Errorf("disk_block_size: Virtual machine requires disk block size <= %v MB, but defined: %v",
			MaxDiskBlockSize, b.config.DiskBlockSize)
	}

	return nil
}

func (b *Builder) checkRamSize() error {
	if b.config.RamSize == 0 {
		b.config.RamSize = DefaultRamSize
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", b.config.RamSize))

	if b.config.RamSize < MinRamSize {
		return fmt.Errorf("memory: Virtual machine requires memory size >= %v MB, but defined: %v",
			MinRamSize, b.config.RamSize)
	} else if b.config.RamSize > MaxRamSize {
		return fmt.Errorf("memory: Virtual machine requires memory size <= %v MB, but defined: %v",
			MaxRamSize, b.config.RamSize)
	}

	return nil
}

func (b *Builder) checkHostAvailableMemory() string {
	powershellAvailable, _, _ := powershell.IsPowershellAvailable()

	if powershellAvailable {
		freeMB := powershell.GetHostAvailableMemory()

		if (freeMB - float64(b.config.RamSize)) < LowRam {
			return fmt.Sprintf("Hyper-V might fail to create a VM if there is not enough free memory in the system.")
		}
	}

	return ""
}

func (b *Builder) detectSwitchName() string {
	powershellAvailable, _, _ := powershell.IsPowershellAvailable()

	if powershellAvailable {
		// no switch name, try to get one attached to a online network adapter
		onlineSwitchName, err := hyperv.GetExternalOnlineVirtualSwitch()
		if onlineSwitchName != "" && err == nil {
			return onlineSwitchName
		}
	}

	return fmt.Sprintf("packer-%s", b.config.PackerBuildName)
}
