package null

import (
	"fmt"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	SSHUsername       string `mapstructure:"ssh_username"`
	SSHPassword       string `mapstructure:"ssh_password"`
	SSHPrivateKeyFile string `mapstructure:"ssh_private_key_file"`
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config

	err := config.Decode(&c, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	if c.Port == 0 {
		c.Port = 22
	}

	var errs *packer.MultiError
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

	return &c, nil, nil
}
