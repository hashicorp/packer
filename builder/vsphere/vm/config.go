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

	BootCommand            []string `mapstructure:"boot_command"`
	Cpu                    uint     `mapstructure:"cpu"`
	MemSize                uint     `mapstructure:"mem_size"`
	NetworkAdapter         string   `mapstructure:"network_adapter"`
	NetworkName            string   `mapstructure:"network_name"`
	SourceVMName           string   `mapstructure:"source_vm"`
	RemoteSourceFolder     string   `mapstructure:"source_folder"`
	RemoteSourceDatacenter string   `mapstructure:"source_datacenter"`
	DiskThick              bool     `mapstructure:"disk_thick"`

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

	if c.SourceVMName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_vm is blank, but is required"))
	}

	if c.RemoteSourceDatacenter == "" {
		c.RemoteSourceDatacenter = "ha-datacenter"
	}

	//For cloning Cpu and Memsize == 0 reuse the same value than for the source VM
	// NetworkName and NetworkAdapter can be empty to reuse configuration from the source VM

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
