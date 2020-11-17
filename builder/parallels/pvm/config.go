//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package pvm

import (
	"fmt"
	"os"

	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/commonsteps"
	"github.com/hashicorp/packer/packer-plugin-sdk/shutdowncommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig                 `mapstructure:",squash"`
	commonsteps.FloppyConfig            `mapstructure:",squash"`
	parallelscommon.OutputConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlConfig        `mapstructure:",squash"`
	parallelscommon.PrlctlPostConfig    `mapstructure:",squash"`
	parallelscommon.PrlctlVersionConfig `mapstructure:",squash"`
	parallelscommon.SSHConfig           `mapstructure:",squash"`
	shutdowncommand.ShutdownConfig      `mapstructure:",squash"`
	bootcommand.BootConfig              `mapstructure:",squash"`
	parallelscommon.ToolsConfig         `mapstructure:",squash"`
	// The path to a PVM directory that acts as the source
	// of this build.
	SourcePath string `mapstructure:"source_path" required:"true"`
	// Virtual disk image is compacted at the end of
	// the build process using prl_disk_tool utility (except for the case that
	// disk_type is set to plain). In certain rare cases, this might corrupt
	// the resulting disk image. If you find this to be the case, you can disable
	// compaction using this configuration value.
	SkipCompaction bool `mapstructure:"skip_compaction" required:"false"`
	// This is the name of the PVM directory for the new
	// virtual machine, without the file extension. By default this is
	// "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
	VMName string `mapstructure:"vm_name" required:"false"`
	// If this is "false" the MAC address of the first
	// NIC will reused when imported else a new MAC address will be generated
	// by Parallels. Defaults to "false".
	ReassignMAC bool `mapstructure:"reassign_mac" required:"false"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         parallelscommon.BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"prlctl",
				"prlctl_post",
				"parallels_tools_guest_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", c.PackerBuildName)
	}

	// Prepare the errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.PrlctlConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.PrlctlPostConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.PrlctlVersionConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ToolsConfig.Prepare(&c.ctx)...)

	if c.SourcePath == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is required"))
	} else {
		if _, err := os.Stat(c.SourcePath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("source_path is invalid: %s", err))
		}
	}

	// Warnings
	var warnings []string
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}
