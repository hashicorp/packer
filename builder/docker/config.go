package docker

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	Commit     bool
	ExportPath string `mapstructure:"export_path"`
	Image      string
	Pull       bool
	RunCommand []string `mapstructure:"run_command"`
	Volumes    map[string]string

	Login         bool
	LoginEmail    string `mapstructure:"login_email"`
	LoginUsername string `mapstructure:"login_username"`
	LoginPassword string `mapstructure:"login_password"`
	LoginServer   string `mapstructure:"login_server"`
	Pty           bool

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	if len(c.RunCommand) == 0 {
		c.RunCommand = []string{
			"-d", "-i", "-t",
			"{{.Image}}",
			"/bin/bash",
		}
	}

	// Default Pull if it wasn't set
	hasPull := false
	for _, k := range md.Keys {
		if k == "Pull" {
			hasPull = true
			break
		}
	}

	if !hasPull {
		c.Pull = true
	}

	// Default to the normal Docker type
	if c.Comm.Type == "" {
		c.Comm.Type = "docker"
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	if c.Image == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("image must be specified"))
	}

	if c.ExportPath != "" && c.Commit {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("both commit and export_path cannot be set"))
	}

	if c.ExportPath != "" {
		if fi, err := os.Stat(c.ExportPath); err == nil && fi.IsDir() {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf(
				"export_path must be a file, not a directory"))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
