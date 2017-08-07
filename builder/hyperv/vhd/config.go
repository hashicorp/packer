package vhd

import (
	"errors"
	"fmt"

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

	// The size, in megabytes, of the hard disk to create for the VM.
	// By default, this is 130048 (about 127 GB).
	DiskSize uint `mapstructure:"disk_size"`
	// The size, in megabytes, of the computer memory in the VM.
	// By default, this is 1024 (about 1 GB).
	RamSize uint `mapstructure:"ram_size"`

	// Should integration services iso be mounted
	GuestAdditionsMode string `mapstructure:"guest_additions_mode"`

	// The path to the integration services iso
	GuestAdditionsPath string `mapstructure:"guest_additions_path"`

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
	if c.Cpu < 1 {
		c.Cpu = 1
	}
	if c.Generation != 2 {
		c.Generation = 1
	}

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)

	err = b.checkDiskSize()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	err = b.checkRamSize()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	////////////
	FINISH ME
	///////////




	if c.Generation == 2 {
		if len(c.FloppyFiles) > 0 || len(c.FloppyDirectories) > 0 {
			err = errors.New("Generation 2 vms don't support floppy drives. Use ISO image instead.")
			errs = packer.MultiErrorAppend(errs, err)
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
