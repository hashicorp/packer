//go:generate struct-markdown

package common

import (
	"fmt"
	"log"
	"os"
	"strings"

	powershell "github.com/hashicorp/packer/builder/hyperv/common/powershell"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell/hyperv"
	"github.com/hashicorp/packer/common"
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

type CommonConfig struct {
	common.FloppyConfig `mapstructure:",squash"`
	common.CDConfig     `mapstructure:",squash"`
	// The block size of the VHD to be created.
	// Recommended disk block size for Linux hyper-v guests is 1 MiB. This
	// defaults to "32" MiB.
	DiskBlockSize uint `mapstructure:"disk_block_size" required:"false"`
	// The amount, in megabytes, of RAM to assign to the
	// VM. By default, this is 1 GB.
	RamSize uint `mapstructure:"memory" required:"false"`
	// A list of ISO paths to
	// attach to a VM when it is booted. This is most useful for unattended
	// Windows installs, which look for an Autounattend.xml file on removable
	// media. By default, no secondary ISO will be attached.
	SecondaryDvdImages []string `mapstructure:"secondary_iso_images" required:"false"`
	// The size or sizes of any
	// additional hard disks for the VM in megabytes. If this is not specified
	// then the VM will only contain a primary hard disk. Additional drives
	// will be attached to the SCSI interface only. The builder uses
	// expandable rather than fixed-size virtual hard disks, so the actual
	// file representing the disk will not use the full size unless it is
	// full.
	AdditionalDiskSize []uint `mapstructure:"disk_additional_size" required:"false"`
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
	SwitchName string `mapstructure:"switch_name" required:"false"`
	// This is the VLAN of the virtual switch's
	// network card. By default none is set. If none is set then a VLAN is not
	// set on the switch's network card. If this value is set it should match
	// the VLAN specified in by vlan_id.
	SwitchVlanId string `mapstructure:"switch_vlan_id" required:"false"`
	// This allows a specific MAC address to be used on
	// the default virtual network card. The MAC address must be a string with
	// no delimiters, for example "0000deadbeef".
	MacAddress string `mapstructure:"mac_address" required:"false"`
	// This is the VLAN of the virtual machine's network
	// card for the new virtual machine. By default none is set. If none is set
	// then VLANs are not set on the virtual machine's network card.
	VlanId string `mapstructure:"vlan_id" required:"false"`
	// The number of CPUs the virtual machine should use. If
	// this isn't specified, the default is 1 CPU.
	Cpu uint `mapstructure:"cpus" required:"false"`
	// The Hyper-V generation for the virtual machine. By
	// default, this is 1. Generation 2 Hyper-V virtual machines do not support
	// floppy drives. In this scenario use secondary_iso_images instead. Hard
	// drives and DVD drives will also be SCSI and not IDE.
	Generation uint `mapstructure:"generation" required:"false"`
	// If true enable MAC address spoofing
	// for the virtual machine. This defaults to false.
	EnableMacSpoofing bool `mapstructure:"enable_mac_spoofing" required:"false"`
	// If true enable dynamic memory for
	// the virtual machine. This defaults to false.
	EnableDynamicMemory bool `mapstructure:"enable_dynamic_memory" required:"false"`
	// If true enable secure boot for the
	// virtual machine. This defaults to false. See secure_boot_template
	// below for additional settings.
	EnableSecureBoot bool `mapstructure:"enable_secure_boot" required:"false"`
	// The secure boot template to be
	// configured. Valid values are "MicrosoftWindows" (Windows) or
	// "MicrosoftUEFICertificateAuthority" (Linux). This only takes effect if
	// enable_secure_boot is set to "true". This defaults to "MicrosoftWindows".
	SecureBootTemplate string `mapstructure:"secure_boot_template" required:"false"`
	// If true enable
	// virtualization extensions for the virtual machine. This defaults to
	// false. For nested virtualization you need to enable MAC spoofing,
	// disable dynamic memory and have at least 4GB of RAM assigned to the
	// virtual machine.
	EnableVirtualizationExtensions bool `mapstructure:"enable_virtualization_extensions" required:"false"`
	// The location under which Packer will create a directory to house all the
	// VM files and folders during the build. By default `%TEMP%` is used
	// which, for most systems, will evaluate to
	// `%USERPROFILE%/AppData/Local/Temp`.
	//
	// The build directory housed under `temp_path` will have a name similar to
	// `packerhv1234567`. The seven digit number at the end of the name is
	// automatically generated by Packer to ensure the directory name is
	// unique.
	TempPath string `mapstructure:"temp_path" required:"false"`
	// This allows you to set the vm version when calling New-VM to generate
	// the vm.
	Version string `mapstructure:"configuration_version" required:"false"`
	// If "true", Packer will not delete the VM from
	// The Hyper-V manager.
	KeepRegistered bool `mapstructure:"keep_registered" required:"false"`

	Communicator string `mapstructure:"communicator"`
	// If true skip compacting the hard disk for
	// the virtual machine when exporting. This defaults to false.
	SkipCompaction bool `mapstructure:"skip_compaction" required:"false"`
	// If true Packer will skip the export of the VM.
	// If you are interested only in the VHD/VHDX files, you can enable this
	// option. The resulting VHD/VHDX file will be output to
	// <output_directory>/Virtual Hard Disks. By default this option is false
	// and Packer will export the VM to output_directory.
	SkipExport bool `mapstructure:"skip_export" required:"false"`
	// Packer defaults to building Hyper-V virtual
	// machines by launching a GUI that shows the console of the machine being
	// built. When this value is set to true, the machine will start without a
	// console.
	Headless bool `mapstructure:"headless" required:"false"`
	// When configured, determines the device or device type that is given preferential
	// treatment when choosing a boot device.
	//
	// For Generation 1:
	//   - `IDE`
	//   - `CD` *or* `DVD`
	//   - `Floppy`
	//   - `NET`
	//
	// For Generation 2:
	//   - `IDE:x:y`
	//   - `SCSI:x:y`
	//   - `CD` *or* `DVD`
	//   - `NET`
	FirstBootDevice string `mapstructure:"first_boot_device" required:"false"`
	// When configured, the boot order determines the order of the devices
	// from which to boot.
	//
	// The device name must be in the form of `SCSI:x:y`, for example,
	// to boot from the first scsi device use `SCSI:0:0`.
	//
	// **NB** You should also set `first_boot_device` (e.g. `DVD`).
	//
	// **NB** Although the VM will have this initial boot order, the OS can
	// change it, for example, Ubuntu 18.04 will modify the boot order to
	// include itself as the first boot option.
	//
	// **NB** This only works for Generation 2 machines.
	BootOrder []string `mapstructure:"boot_order" required:"false"`
}

func (c *CommonConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) ([]error, []string) {
	// Accumulate any errors and warns
	var errs []error
	var warns []string

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s", pc.PackerBuildName)
		log.Println(fmt.Sprintf("%s: %v", "VMName", c.VMName))
	}

	if c.SwitchName == "" {
		c.SwitchName = c.detectSwitchName(pc.PackerBuildName)
		log.Println(fmt.Sprintf("Using switch %s", c.SwitchName))
	}

	if c.Generation < 1 || c.Generation > 2 {
		c.Generation = 1
	}

	if c.Generation == 2 {
		if len(c.FloppyFiles) > 0 || len(c.FloppyDirectories) > 0 {
			err := fmt.Errorf("Generation 2 vms don't support floppy drives. Use ISO image instead.")
			errs = append(errs, err)
		}
	}

	if len(c.AdditionalDiskSize) > 64 {
		errs = append(errs, fmt.Errorf("VM's currently support a maximum of 64 additional SCSI attached disks."))
	}

	// Errors
	errs = append(errs, c.FloppyConfig.Prepare(ctx)...)
	errs = append(errs, c.CDConfig.Prepare(ctx)...)
	if c.GuestAdditionsMode == "" {
		if c.GuestAdditionsPath != "" {
			c.GuestAdditionsMode = "attach"
		} else {
			c.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

			if _, err := os.Stat(c.GuestAdditionsPath); os.IsNotExist(err) {
				if err != nil {
					c.GuestAdditionsPath = ""
					c.GuestAdditionsMode = "none"
				} else {
					c.GuestAdditionsMode = "attach"
				}
			}
		}
	}

	if c.GuestAdditionsPath == "" && c.GuestAdditionsMode == "attach" {
		c.GuestAdditionsPath = os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

		if _, err := os.Stat(c.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				c.GuestAdditionsPath = ""
			}
		}
	}

	for _, isoPath := range c.SecondaryDvdImages {
		if _, err := os.Stat(isoPath); os.IsNotExist(err) {
			if err != nil {
				errs = append(
					errs, fmt.Errorf("Secondary Dvd image does not exist: %s", err))
			}
		}
	}

	numberOfIsos := len(c.SecondaryDvdImages)

	if c.GuestAdditionsMode == "attach" {
		if _, err := os.Stat(c.GuestAdditionsPath); os.IsNotExist(err) {
			if err != nil {
				errs = append(
					errs, fmt.Errorf("Guest additions iso does not exist: %s", err))
			}
		}

		numberOfIsos = numberOfIsos + 1
	}

	if c.Generation < 2 && numberOfIsos > 2 {
		if c.GuestAdditionsMode == "attach" {
			errs = append(errs, fmt.Errorf("There are only 2 ide controllers available, so "+
				"we can't support guest additions and these secondary dvds: %s",
				strings.Join(c.SecondaryDvdImages, ", ")))
		} else {
			errs = append(errs, fmt.Errorf("There are only 2 ide controllers available, so "+
				"we can't support these secondary dvds: %s",
				strings.Join(c.SecondaryDvdImages, ", ")))
		}
	} else if c.Generation > 1 && len(c.SecondaryDvdImages) > 16 {
		if c.GuestAdditionsMode == "attach" {
			errs = append(errs, fmt.Errorf("There are not enough drive letters available for "+
				"scsi (limited to 16), so we can't support guest additions and these secondary dvds: %s",
				strings.Join(c.SecondaryDvdImages, ", ")))
		} else {
			errs = append(errs, fmt.Errorf("There are not enough drive letters available for "+
				"scsi (limited to 16), so we can't support these secondary dvds: %s",
				strings.Join(c.SecondaryDvdImages, ", ")))
		}
	}

	if c.EnableVirtualizationExtensions {
		hasVirtualMachineVirtualizationExtensions, err := powershell.HasVirtualMachineVirtualizationExtensions()
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed detecting virtual machine virtualization "+
				"extensions support: %s", err))
		} else {
			if !hasVirtualMachineVirtualizationExtensions {
				errs = append(errs, fmt.Errorf("This version of Hyper-V does not support "+
					"virtual machine virtualization extension. Please use Windows 10 or Windows Server 2016 "+
					"or newer."))
			}
		}
	}

	if c.FirstBootDevice != "" {
		_, _, _, err := ParseBootDeviceIdentifier(c.FirstBootDevice, c.Generation)
		if err != nil {
			errs = append(errs, fmt.Errorf("first_boot_device: %s", err))
		}
	}

	if c.EnableVirtualizationExtensions {
		if c.EnableDynamicMemory {
			warning := fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, " +
				"dynamic memory should not be allowed.")
			warns = Appendwarns(warns, warning)
		}

		if !c.EnableMacSpoofing {
			warning := fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, " +
				"mac spoofing should be allowed.")
			warns = Appendwarns(warns, warning)
		}

		if c.RamSize < MinNestedVirtualizationRamSize {
			warning := fmt.Sprintf("For nested virtualization, when virtualization extension is enabled, " +
				"there should be 4GB or more memory set for the vm, otherwise Hyper-V may fail to start " +
				"any nested VMs.")
			warns = Appendwarns(warns, warning)
		}
	}

	if c.SwitchVlanId != "" {
		if c.SwitchVlanId != c.VlanId {
			warning := fmt.Sprintf("Switch network adaptor vlan should match virtual machine network adaptor " +
				"vlan. The switch will not be able to see traffic from the VM.")
			warns = Appendwarns(warns, warning)
		}
	}

	err := c.checkDiskBlockSize()
	if err != nil {
		errs = append(errs, err)
	}
	err = c.checkRamSize()
	if err != nil {
		errs = append(errs, err)
	}

	// warns
	warning := c.checkHostAvailableMemory()
	if warning != "" {
		warns = Appendwarns(warns, warning)
	}

	if errs != nil && len(errs) > 0 {
		return errs, warns
	}

	return nil, warns
}

func (c *CommonConfig) checkDiskBlockSize() error {
	if c.DiskBlockSize == 0 {
		c.DiskBlockSize = DefaultDiskBlockSize
	}

	log.Println(fmt.Sprintf("%s: %v", "DiskBlockSize", c.DiskBlockSize))

	if c.DiskBlockSize < MinDiskBlockSize {
		return fmt.Errorf("disk_block_size: Virtual machine requires disk block size >= %v MB, but defined: %v",
			MinDiskBlockSize, c.DiskBlockSize)
	} else if c.DiskBlockSize > MaxDiskBlockSize {
		return fmt.Errorf("disk_block_size: Virtual machine requires disk block size <= %v MB, but defined: %v",
			MaxDiskBlockSize, c.DiskBlockSize)
	}

	return nil
}

func (c *CommonConfig) checkHostAvailableMemory() string {
	powershellAvailable, _, _ := powershell.IsPowershellAvailable()

	if powershellAvailable {
		freeMB := powershell.GetHostAvailableMemory()

		if (freeMB - float64(c.RamSize)) < LowRam {
			return fmt.Sprintf("Hyper-V might fail to create a VM if there is not enough free memory in the system.")
		}
	}

	return ""
}

func (c *CommonConfig) checkRamSize() error {
	if c.RamSize == 0 {
		c.RamSize = DefaultRamSize
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", c.RamSize))

	if c.RamSize < MinRamSize {
		return fmt.Errorf("memory: Virtual machine requires memory size >= %v MB, but defined: %v",
			MinRamSize, c.RamSize)
	} else if c.RamSize > MaxRamSize {
		return fmt.Errorf("memory: Virtual machine requires memory size <= %v MB, but defined: %v",
			MaxRamSize, c.RamSize)
	}

	return nil
}

func (c *CommonConfig) detectSwitchName(buildName string) string {
	powershellAvailable, _, _ := powershell.IsPowershellAvailable()

	if powershellAvailable {
		// no switch name, try to get one attached to a online network adapter
		onlineSwitchName, err := hyperv.GetExternalOnlineVirtualSwitch()
		if onlineSwitchName != "" && err == nil {
			return onlineSwitchName
		}
	}

	return fmt.Sprintf("packer-%s", buildName)
}

func Appendwarns(slice []string, data ...string) []string {
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
