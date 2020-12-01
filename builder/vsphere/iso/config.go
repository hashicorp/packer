//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package iso

import (
	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/helper/communicator"
	packerCommon "github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	packerCommon.PackerConfig `mapstructure:",squash"`
	commonsteps.HTTPConfig    `mapstructure:",squash"`
	commonsteps.CDConfig      `mapstructure:",squash"`

	common.ConnectConfig      `mapstructure:",squash"`
	CreateConfig              `mapstructure:",squash"`
	common.LocationConfig     `mapstructure:",squash"`
	common.HardwareConfig     `mapstructure:",squash"`
	common.ConfigParamsConfig `mapstructure:",squash"`

	commonsteps.ISOConfig `mapstructure:",squash"`

	common.CDRomConfig       `mapstructure:",squash"`
	common.RemoveCDRomConfig `mapstructure:",squash"`
	common.FloppyConfig      `mapstructure:",squash"`
	common.RunConfig         `mapstructure:",squash"`
	common.BootConfig        `mapstructure:",squash"`
	common.WaitIpConfig      `mapstructure:",squash"`
	Comm                     communicator.Config `mapstructure:",squash"`

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
		PluginType:         common.BuilderId,
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
	errs := new(packersdk.MultiError)

	if c.ISOUrls != nil || c.RawSingleISOUrl != "" {
		isoWarnings, isoErrs := c.ISOConfig.Prepare(&c.ctx)
		warnings = append(warnings, isoWarnings...)
		errs = packersdk.MultiErrorAppend(errs, isoErrs...)
	}

	errs = packersdk.MultiErrorAppend(errs, c.ConnectConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.CreateConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.LocationConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.HardwareConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)

	errs = packersdk.MultiErrorAppend(errs, c.CDRomConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.CDConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.WaitIpConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)

	shutdownWarnings, shutdownErrs := c.ShutdownConfig.Prepare(c.Comm)
	warnings = append(warnings, shutdownWarnings...)
	errs = packersdk.MultiErrorAppend(errs, shutdownErrs...)

	if c.Export != nil {
		errs = packersdk.MultiErrorAppend(errs, c.Export.Prepare(&c.ctx, &c.LocationConfig, &c.PackerConfig)...)
	}
	if c.ContentLibraryDestinationConfig != nil {
		errs = packersdk.MultiErrorAppend(errs, c.ContentLibraryDestinationConfig.Prepare(&c.LocationConfig)...)
	}

	if len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}
