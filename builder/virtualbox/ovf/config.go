package ovf

import (
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// Config is the configuration structure for the builder.
type Config struct {
	common.PackerConfig         `mapstructure:",squash"`
	vboxcommon.FloppyConfig     `mapstructure:",squash"`
	vboxcommon.OutputConfig     `mapstructure:",squash"`
	vboxcommon.SSHConfig        `mapstructure:",squash"`
	vboxcommon.VBoxManageConfig `mapstructure:",squash"`

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

	// Prepare the errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, c.FloppyConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.OutputConfig.Prepare(c.tpl, &c.PackerConfig)...)
	errs = packer.MultiErrorAppend(errs, c.SSHConfig.Prepare(c.tpl)...)
	errs = packer.MultiErrorAppend(errs, c.VBoxManageConfig.Prepare(c.tpl)...)

	// Check for any errors.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
