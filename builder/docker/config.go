package docker

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

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

	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"export_path":    &c.ExportPath,
		"image":          &c.Image,
		"login_email":    &c.LoginEmail,
		"login_username": &c.LoginUsername,
		"login_password": &c.LoginPassword,
		"login_server":   &c.LoginServer,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	for k, v := range c.Volumes {
		var err error
		v, err = c.tpl.Process(v, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing volumes[%s]: %s", k, err))
		}

		c.Volumes[k] = v
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("image must be specified"))
	}

	if c.ExportPath != "" && c.Commit {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("both commit and export_path cannot be set"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
