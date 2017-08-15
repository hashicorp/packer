package vhd

import (
	"errors"
	"fmt"
	"log"

	hypervcommon "github.com/hashicorp/packer/builder/hyperv/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/powershell"
	"github.com/hashicorp/packer/common/powershell/hyperv"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig               `mapstructure:",squash"`
	common.HTTPConfig                 `mapstructure:",squash"`
	common.ISOConfig                  `mapstructure:",squash"`
	common.FloppyConfig               `mapstructure:",squash"`
	hypervcommon.OutputConfig         `mapstructure:",squash"`
	hypervcommon.SSHConfig            `mapstructure:",squash"`
	hypervcommon.RunConfig            `mapstructure:",squash"`
	hypervcommon.ShutdownConfig       `mapstructure:",squash"`
	hypervcommon.GuestAdditionsConfig `mapstructure:",squash"`
	hypervcommon.SizeConfig           `mapstructure:",squash"`

	BootCommand                    []string `mapstructure:"boot_command"`
	Checksum                       string   `mapstructure:"checksum"`
	ChecksumType                   string   `mapstructure:"checksum_type"`
	SwitchName                     string   `mapstructure:"switch_name"`
	SwitchVlanId                   string   `mapstructure:"switch_vlan_id"`
	VlanId                         string   `mapstructure:"vlan_id"`
	Cpu                            uint     `mapstructure:"cpu"`
	Generation                     uint     `mapstructure:"generation"`
	EnableMacSpoofing              bool     `mapstructure:"enable_mac_spoofing"`
	EnableDynamicMemory            bool     `mapstructure:"enable_dynamic_memory"`
	EnableSecureBoot               bool     `mapstructure:"enable_secure_boot"`
	EnableVirtualizationExtensions bool     `mapstructure:"enable_virtualization_extensions"`
	SourcePath                     string   `mapstructure:"source_path"`
	TempPath                       string   `mapstructure:"temp_path"`
	VMName                         string   `mapstructure:"vm_name"`

	Communicator string `mapstructure:"communicator"`

	SkipCompaction bool `mapstructure:"skip_compaction"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults and clamping
	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", c.PackerBuildName)
	}

	if c.SwitchName == "" {
		c.SwitchName = c.detectSwitchName()
	}
	log.Println(fmt.Sprintf("Using switch %s", c.SwitchName))

	if c.Cpu < 1 {
		c.Cpu = 1
	}
	if c.Generation != 2 {
		c.Generation = 1
	}

	// Accumulate errors
	var errs *packer.MultiError

	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SizeConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.GuestAdditionsConfig.Prepare(&c.ctx, []string{}, c.Generation)...)

	if c.Generation == 2 {
		if len(c.FloppyFiles) > 0 || len(c.FloppyDirectories) > 0 {
			err = errors.New("Generation 2 vms don't support floppy drives.")
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.EnableVirtualizationExtensions {
		if hasVMVirtExts, err := powershell.HasVirtualMachineVirtualizationExtensions(); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed detecting virtual machine virtualization extensions support: %s", err))
		} else if !hasVMVirtExts {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("This version of Hyper-V does not support virtual machine virtualization extension. Please use Windows 10 or Windows Server 2016 or newer."))
		}
	}

	// Accumulate warnings
	warnings := make([]string, 0)

	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	if warn := c.SizeConfig.ValidateAvailable(); warn != "" {
		warnings = append(warnings, warn)
	}

	if c.EnableVirtualizationExtensions {
		if c.EnableDynamicMemory {
			warnings = append(warnings,
				"For nested virtualization, when virtualization extension is enabled, dynamic memory should not be allowed.")
		}

		if !c.EnableMacSpoofing {
			warnings = append(warnings,
				"For nested virtualization, when virtualization extension is enabled, mac spoofing should be allowed.")
		}

		if warn := c.SizeConfig.ValidateMinimum(); warn != "" {
			warnings = append(warnings, warn)
		}
	}

	if c.SwitchVlanId != "" {
		if c.SwitchVlanId != c.VlanId {
			warnings = append(warnings,
				"Switch network adaptor vlan should match virtual machine network adaptor vlan. The switch will not be able to see traffic from the VM.")
		}
	}

	return c, warnings, errs
}

func (c *Config) detectSwitchName() string {
	powershellAvailable, _, _ := powershell.IsPowershellAvailable()

	if powershellAvailable {
		// no switch name, try to get one attached to a online network adapter
		onlineSwitchName, err := hyperv.GetExternalOnlineVirtualSwitch()
		if onlineSwitchName != "" && err == nil {
			return onlineSwitchName
		}
	}

	return fmt.Sprintf("packer-%s", c.PackerBuildName)
}
