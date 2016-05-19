// +build linux

package lxc

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"time"
	"os"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ConfigFile          string   `mapstructure:"config_file"`
	OutputDir           string   `mapstructure:"output_directory"`
	ContainerName       string   `mapstructure:"container_name"`
	CommandWrapper      string   `mapstructure:"command_wrapper"`
	RawInitTimeout      string   `mapstructure:"init_timeout"`
	Name                string   `mapstructure:"template_name"`
	Parameters          []string `mapstructure:"template_parameters"`
	EnvVars             []string `mapstructure:"template_environment_vars"`
	TargetRunlevel      int      `mapstructure:"target_runlevel"`
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

	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", c.PackerBuildName)
	}

	if c.ContainerName == "" {
		c.ContainerName = fmt.Sprintf("packer-%s", c.PackerBuildName)
	}

	if c.TargetRunlevel == 0 {
		c.TargetRunlevel = 3
	}

	if c.CommandWrapper == "" {
		c.CommandWrapper = "{{.Command}}"
	}

	if c.RawInitTimeout == "" {
		c.RawInitTimeout = "20s"
	}

	c.InitTimeout, err = time.ParseDuration(c.RawInitTimeout)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed parsing init_timeout: %s", err))
	}

	if _, err := os.Stat(c.ConfigFile); os.IsNotExist(err) {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("LXC Config file appears to be missing: %s", c.ConfigFile))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return &c, nil
}
