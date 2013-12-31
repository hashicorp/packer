package ovf

import (
	"fmt"
	"os"

	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig          `mapstructure:",squash"`
	vboxcommon.ExportConfig      `mapstructure:",squash"`
	vboxcommon.FloppyConfig      `mapstructure:",squash"`
	vboxcommon.OutputConfig      `mapstructure:",squash"`
	vboxcommon.RunConfig         `mapstructure:",squash"`
	vboxcommon.SSHConfig         `mapstructure:",squash"`
	vboxcommon.ShutdownConfig    `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig  `mapstructure:",squash"`
	vboxcommon.VBoxVersionConfig `mapstructure:",squash"`

	SourcePath string `mapstructure:"source_path"`
	VMName     string `mapstructure:"vm_name"`

	tpl *packer.ConfigTemplate
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	md, err := common.DecodeConfig(c, raws...)
	if err != nil {
		return nil, nil, err
	}

	c.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, nil, err
	}
	c.tpl.UserVars = c.PackerUserVars

	// Defaults
	if c.VMName == "" {
		c.VMName = fmt.Sprintf("packer-%s-{{timestamp}}", c.PackerBuildName)
	}

	// Prepare the errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, c.ExportConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(c.tpl, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.RunConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.ShutdownConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManageConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxVersionConfig.Prepare(c.tpl)...)

	templates := map[string]*string{
		"source_path": &c.SourcePath,
		"vm_name":     &c.VMName,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

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
		return nil, warnings, errs
	}

	return c, warnings, nil
}
