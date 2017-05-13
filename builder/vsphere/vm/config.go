package vm

import (
	"fmt"

	vspcommon "github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	common.HTTPConfig        `mapstructure:",squash"`
	vspcommon.DriverConfig   `mapstructure:",squash"`
	vspcommon.OutputConfig   `mapstructure:",squash"`
	vspcommon.RunConfig      `mapstructure:",squash"`
	vspcommon.ShutdownConfig `mapstructure:",squash"`
	vspcommon.SSHConfig      `mapstructure:",squash"`
	vspcommon.VMXConfig      `mapstructure:",squash"`
	vspcommon.ExportConfig   `mapstructure:",squash"`

	//TODO: Review this options to provide all needed information for vm.clone
	BootCommand    []string `mapstructure:"boot_command"`
	Cpu            uint     `mapstructure:"cpu"`
	DiskSize       uint     `mapstructure:"disk_size"`
	DiskThick      bool     `mapstructure:"disk_thick"`
	MemSize        uint     `mapstructure:"mem_size"`
	NetworkAdapter string   `mapstructure:"network_adapter"`
	NetworkName    string   `mapstructure:"network_name"`
	SourceVMName   string   `mapstructure:"source_vm"`

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
	shutWarnings, shutErrs := c.ShutdownConfig.Prepare(&c.ctx)
	warnings = append(warnings, shutWarnings...)
	errs = packer.MultiErrorAppend(errs, shutErrs...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.DriverConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VMXConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)

	//TODO: Review this part with options for vm.clone
	if c.SourceVMName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_vm is blank, but is required"))
	}

	//For cloning DiskSize,Cpu and Memsize == 0 reuse the same value than for the source VM
	// NetworkName and NetworkAdapter can be empty to reuse configuration from the source VM

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
