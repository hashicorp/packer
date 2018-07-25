package lxd

import (
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	OutputImage         string            `mapstructure:"output_image"`
	ContainerName       string            `mapstructure:"container_name"`
	CommandWrapper      string            `mapstructure:"command_wrapper"`
	Image               string            `mapstructure:"image"`
	Profile             string            `mapstructure:"profile"`
	InitSleep           string            `mapstructure:"init_sleep"`
	PublishProperties   map[string]string `mapstructure:"publish_properties"`
	LaunchConfig        map[string]string `mapstructure:"launch_config"`

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

	if c.Profile == "" {
		c.Profile = "default"
	}

	// Sadly we have to wait a few seconds for /tmp to be intialized and networking
	// to finish starting. There isn't a great cross platform to check when things are ready.
	if c.InitSleep == "" {
		c.InitSleep = "3"
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return &c, nil
}
