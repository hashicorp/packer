package iso

import (
	vspcommon "github.com/mitchellh/packer/builder/vsphere/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	common.ISOConfig         `mapstructure:",squash"`
	common.HTTPConfig        `mapstructure:",squash"`
	common.FloppyConfig      `mapstructure:",squash"`
	vspcommon.DriverConfig   `mapstructure:",squash"`
	vspcommon.OutputConfig   `mapstructure:",squash"`
	vspcommon.RunConfig      `mapstructure:",squash"`
	vspcommon.ShutdownConfig `mapstructure:",squash"`
	vspcommon.SSHConfig      `mapstructure:",squash"`
	vspcommon.VMXConfig      `mapstructure:",squash"`
	vspcommon.ExportConfig   `mapstructure:",squash"`
	BootCommand              []string `mapstructure:"boot_command"`
	Cpu                      uint     `mapstructure:"cpu"`
	DiskSize                 uint     `mapstructure:"disk_size"`
	DiskThick                bool     `mapstructure:"disk_thick"`
	// List of possible values in https://github.com/vmware/pyvmomi/blob/master/docs/vim/vm/GuestOsDescriptor/GuestOsIdentifier.rst
	GuestOSType    string `mapstructure:"guest_os_type"`
	MemSize        uint   `mapstructure:"mem_size"`
	NetworkAdapter string `mapstructure:"network_adapter"`
	NetworkName    string `mapstructure:"network_name"`

	CommConfig communicator.Config `mapstructure:",squash"`

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

	// Accumulate any errors and warnings
	var errs *packer.MultiError
	warnings := make([]string, 0)

	isoWarnings, isoErrs := c.ISOConfig.Prepare(&c.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)
	shutWarnings, shutErrs := c.ShutdownConfig.Prepare(&c.ctx)
	warnings = append(warnings, shutWarnings...)
	errs = packer.MultiErrorAppend(errs, shutErrs...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.DriverConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VMXConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)

	if c.DiskSize == 0 {
		c.DiskSize = 40000
	}

	if c.Cpu == 0 {
		c.Cpu = 1
	}

	if c.MemSize == 0 {
		c.MemSize = 1024
	}

	if c.NetworkAdapter == "" {
		c.NetworkAdapter = "e1000"
	}

	if c.NetworkName == "" {
		c.NetworkName = "VM Network"
	}

	if c.GuestOSType == "" {
		c.GuestOSType = "otherGuest"
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
