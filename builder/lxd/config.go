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
	OutputImage         string `mapstructure:"output_image"`
	ContainerName       string `mapstructure:"container_name"`
	CommandWrapper      string `mapstructure:"command_wrapper"`
	Image               string `mapstructure:"image"`
	InitTimeout         time.Duration

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

	if c.Image == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("`image` is a required parameter for LXD. Please specify an image by alias or fingerprint. e.g. `ubuntu-daily:x`"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return &c, nil
}
