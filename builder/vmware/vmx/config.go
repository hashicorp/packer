//go:generate struct-markdown

package vmx

import (
	"fmt"
	"os"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	common.HTTPConfig        `mapstructure:",squash"`
	common.FloppyConfig      `mapstructure:",squash"`
	bootcommand.VNCConfig    `mapstructure:",squash"`
	vmwcommon.DriverConfig   `mapstructure:",squash"`
	vmwcommon.OutputConfig   `mapstructure:",squash"`
	vmwcommon.RunConfig      `mapstructure:",squash"`
	vmwcommon.ShutdownConfig `mapstructure:",squash"`
	vmwcommon.SSHConfig      `mapstructure:",squash"`
	vmwcommon.ToolsConfig    `mapstructure:",squash"`
	vmwcommon.VMXConfig      `mapstructure:",squash"`
	vmwcommon.ExportConfig   `mapstructure:",squash"`
	// By default Packer creates a 'full' clone of
    // the virtual machine specified in source_path. The resultant virtual
    // machine is fully independant from the parent it was cloned from.
	Linked     bool   `mapstructure:"linked" required:"false"`
	// The type of remote machine that will be used to
    // build this VM rather than a local desktop product. The only value accepted
    // for this currently is esx5. If this is not set, a desktop product will
    // be used. By default, this is not set.
	RemoteType string `mapstructure:"remote_type" required:"false"`
	// Path to the source VMX file to clone. If
    // remote_type is enabled then this specifies a path on the remote_host.
	SourcePath string `mapstructure:"source_path" required:"true"`
	// This is the name of the VMX file for the new virtual
    // machine, without the file extension. By default this is packer-BUILDNAME,
    // where "BUILDNAME" is the name of the build.
	VMName     string `mapstructure:"vm_name" required:"false"`

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
				"tools_upload_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	if c.VMName == "" {
		c.VMName = fmt.Sprintf(
			"packer-%s-%d", c.PackerBuildName, interpolate.InitTime.Unix())
	}

	// Prepare the errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.DriverConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(&c.ctx, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ToolsConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VMXConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.VNCConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(&c.ctx)...)

	if c.RemoteType == "" {
		if c.SourcePath == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("source_path is blank, but is required"))
		} else {
			if _, err := os.Stat(c.SourcePath); err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("source_path is invalid: %s", err))
			}
		}
	} else {
		// Remote configuration validation
		if c.RemoteHost == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("remote_host must be specified"))
		}

		if c.RemoteType != "esx5" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Only 'esx5' value is accepted for remote_type"))
		}
	}

	err = c.DriverConfig.Validate(c.SkipExport)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if c.Format != "" {
		if c.RemoteType != "esx5" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("format is only valid when remote_type=esx5"))
		}
	} else {
		c.Format = "ovf"
	}

	if !(c.Format == "ova" || c.Format == "ovf" || c.Format == "vmx") {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("format must be one of ova, ovf, or vmx"))
	}

	// Warnings
	var warnings []string
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	if c.Headless && c.DisableVNC {
		warnings = append(warnings,
			"Headless mode uses VNC to retrieve output. Since VNC has been disabled,\n"+
				"you won't be able to see any output.")
	}

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
