package lxd

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"time"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	///ConfigFile          string   `mapstructure:"config_file"`
	OutputImage    string `mapstructure:"output_image"`
	ContainerName  string `mapstructure:"container_name"`
	CommandWrapper string `mapstructure:"command_wrapper"`
	RawInitTimeout string `mapstructure:"init_timeout"`
	Image          string `mapstructure:"image"`
	//EnvVars             []string `mapstructure:"template_environment_vars"`
	InitTimeout time.Duration

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, error) {
	var c Config

	var md mapstructure.Metadata
	err := config.Decode(&c, &config.DecodeOpts{
		Metadata:    &md,
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors
	var errs *packer.MultiError

	if c.ContainerName == "" {
		c.ContainerName = fmt.Sprintf("packer-%s", c.PackerBuildName)
	}

	if c.OutputImage == "" {
		c.OutputImage = c.ContainerName
	}

	if c.CommandWrapper == "" {
		c.CommandWrapper = "{{.Command}}"
	}

	if c.RawInitTimeout == "" {
		c.RawInitTimeout = "20s"
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("`image` is a required parameter for LXD."))
	}

	c.InitTimeout, err = time.ParseDuration(c.RawInitTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed parsing init_timeout: %s", err))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return &c, nil
}
