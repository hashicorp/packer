//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package iso

import (
	"github.com/hashicorp/packer/builder/vsphere/common"
	packerCommon "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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

	CDRomConfig         `mapstructure:",squash"`
	RemoveCDRomConfig   `mapstructure:",squash"`
	FloppyConfig        `mapstructure:",squash"`
	common.RunConfig    `mapstructure:",squash"`
	BootConfig          `mapstructure:",squash"`
	common.WaitIpConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	common.ShutdownConfig `mapstructure:",squash"`

	// Create a snapshot when set to `true`, so the VM can be used as a base
	// for linked clones. Defaults to `false`.
	CreateSnapshot bool `mapstructure:"create_snapshot"`
	// Convert VM to a template. Defaults to `false`.
	ConvertToTemplate bool `mapstructure:"convert_to_template"`
	// Configuration for exporting VM to an ovf file.
	// The VM will not be exported if no [Export Configuration](#export-configuration) is specified.
	Export *common.ExportConfig `mapstructure:"export"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
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
		return nil, err
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
	errs = packer.MultiErrorAppend(errs, c.WaitIpConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare()...)
	if c.Export != nil {
		errs = packer.MultiErrorAppend(errs, c.Export.Prepare(&c.ctx, &c.LocationConfig, &c.PackerConfig)...)
	}

	if len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}
