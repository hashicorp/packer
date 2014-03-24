package null

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	SSHUsername       string `mapstructure:"ssh_username"`
	SSHPassword       string `mapstructure:"ssh_password"`
	SSHPrivateKeyFile string `mapstructure:"ssh_private_key_file"`

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
	if c.Port == 0 {
		c.Port = 22
	}
	// (none so far)

	errs := common.CheckUnusedConfig(md)

	if c.Host == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("host must be specified"))
	}

	if c.SSHUsername == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("ssh_username must be specified"))
	}

	if c.SSHPassword == "" && c.SSHPrivateKeyFile == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("one of ssh_password and ssh_private_key_file must be specified"))
	}

	if c.SSHPassword != "" && c.SSHPrivateKeyFile != "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("only one of ssh_password and ssh_private_key_file must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
