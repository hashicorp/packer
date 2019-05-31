//go:generate struct-markdown

package lxc

import (
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// The path to the lxc configuration file.
	ConfigFile          string   `mapstructure:"config_file" required:"true"`
	// The directory in which to save the exported
    // tar.gz. Defaults to output-<BuildName> in the current directory.
	OutputDir           string   `mapstructure:"output_directory" required:"false"`
	// The name of the LXC container. Usually stored
    // in /var/lib/lxc/containers/<container_name>. Defaults to
    // packer-<BuildName>.
	ContainerName       string   `mapstructure:"container_name" required:"false"`
	// Allows you to specify a wrapper command, such
    // as ssh so you can execute packer builds on a remote host. Defaults to
    // Empty.
	CommandWrapper      string   `mapstructure:"command_wrapper" required:"false"`
	// The timeout in seconds to wait for the the
    // container to start. Defaults to 20 seconds.
	RawInitTimeout      string   `mapstructure:"init_timeout" required:"false"`
	// Options to pass to lxc-create. For
    // instance, you can specify a custom LXC container configuration file with
    // ["-f", "/path/to/lxc.conf"]. Defaults to []. See man 1 lxc-create for
    // available options.
	CreateOptions       []string `mapstructure:"create_options" required:"false"`
	// Options to pass to lxc-start. For
    // instance, you can override parameters from the LXC container configuration
    // file via ["--define", "KEY=VALUE"]. Defaults to []. See
    // man 1 lxc-start for available options.
	StartOptions        []string `mapstructure:"start_options" required:"false"`
	// Options to pass to lxc-attach. For
    // instance, you can prevent the container from inheriting the host machine's
    // environment by specifying ["--clear-env"]. Defaults to []. See
    // man 1 lxc-attach for available options.
	AttachOptions       []string `mapstructure:"attach_options" required:"false"`
	// The LXC template name to use.
	Name                string   `mapstructure:"template_name" required:"true"`
	// Options to pass to the given
    // lxc-template command, usually located in
    // /usr/share/lxc/templates/lxc-<template_name>. Note: This gets passed as
    // ARGV to the template command. Ensure you have an array of strings, as a
    // single string with spaces probably won't work. Defaults to [].
	Parameters          []string `mapstructure:"template_parameters" required:"false"`
	// Environmental variables to
    // use to build the template with.
	EnvVars             []string `mapstructure:"template_environment_vars" required:"true"`
	// The minimum run level to wait for the
    // container to reach. Note some distributions (Ubuntu) simulate run levels
    // and may report 5 rather than 3.
	TargetRunlevel      int      `mapstructure:"target_runlevel" required:"false"`
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
