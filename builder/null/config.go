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

	if c.Port == 0 {
		c.Port = 22
	}

	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"host":                 &c.Host,
		"ssh_username":         &c.SSHUsername,
		"ssh_password":         &c.SSHPassword,
		"ssh_private_key_file": &c.SSHPrivateKeyFile,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

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
