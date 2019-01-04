package iso

import (
	packerCommon "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
)

type Config struct {
	packerCommon.PackerConfig `mapstructure:",squash"`
	packerCommon.HTTPConfig   `mapstructure:",squash"`

	common.ConnectConfig      `mapstructure:",squash"`
	CreateConfig              `mapstructure:",squash"`
	common.LocationConfig     `mapstructure:",squash"`
	common.HardwareConfig     `mapstructure:",squash"`
	common.ConfigParamsConfig `mapstructure:",squash"`

	packerCommon.ISOConfig `mapstructure:",squash"`

	CDRomConfig      `mapstructure:",squash"`
	FloppyConfig     `mapstructure:",squash"`
	common.RunConfig `mapstructure:",squash"`
	BootConfig       `mapstructure:",squash"`
	Comm             communicator.Config `mapstructure:",squash"`

	common.ShutdownConfig `mapstructure:",squash"`

	CreateSnapshot    bool `mapstructure:"create_snapshot"`
	ConvertToTemplate bool `mapstructure:"convert_to_template"`

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

	warnings := make([]string, 0)
	errs := new(packer.MultiError)

	if c.ISOUrls != nil {
		isoWarnings, isoErrs := c.ISOConfig.Prepare(&c.ctx)
		warnings = append(warnings, isoWarnings...)
		errs = packer.MultiErrorAppend(errs, isoErrs...)
	}

	errs = packer.MultiErrorAppend(errs, c.ConnectConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.CreateConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.LocationConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.HardwareConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)

	errs = packer.MultiErrorAppend(errs, c.CDRomConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare()...)

	if len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
