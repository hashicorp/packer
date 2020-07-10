//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package clone

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
	CloneConfig               `mapstructure:",squash"`
	common.LocationConfig     `mapstructure:",squash"`
	common.HardwareConfig     `mapstructure:",squash"`
	common.ConfigParamsConfig `mapstructure:",squash"`

	common.RunConfig      `mapstructure:",squash"`
	common.BootConfig     `mapstructure:",squash"`
	common.WaitIpConfig   `mapstructure:",squash"`
	Comm                  communicator.Config `mapstructure:",squash"`
	common.ShutdownConfig `mapstructure:",squash"`

	// Create a snapshot when set to `true`, so the VM can be used as a base
	// for linked clones. Defaults to `false`.
	CreateSnapshot bool `mapstructure:"create_snapshot"`
	// Convert VM to a template. Defaults to `false`.
	ConvertToTemplate bool `mapstructure:"convert_to_template"`
	// Configuration for exporting VM to an ovf file.
	// The VM will not be exported if no [Export Configuration](#export-configuration) is specified.
	Export *common.ExportConfig `mapstructure:"export"`
	// Configuration for importing the VM template to a Content Library.
	// The VM template will not be imported if no [Content Library Import Configuration](#content-library-import-configuration) is specified.
	// The import doesn't work if [convert_to_template](#convert_to_template) is set to true.
	ContentLibraryDestinationConfig *common.ContentLibraryDestinationConfig `mapstructure:"content_library_destination"`

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

	errs := new(packer.MultiError)
	errs = packer.MultiErrorAppend(errs, c.ConnectConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.CloneConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.LocationConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.HardwareConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)

	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.WaitIpConfig.Prepare()...)
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare()...)
	if c.Export != nil {
		errs = packer.MultiErrorAppend(errs, c.Export.Prepare(&c.ctx, &c.LocationConfig, &c.PackerConfig)...)
	}
	if c.ContentLibraryDestinationConfig != nil {
		errs = packer.MultiErrorAppend(errs, c.ContentLibraryDestinationConfig.Prepare(&c.LocationConfig)...)
	}

	if len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}
