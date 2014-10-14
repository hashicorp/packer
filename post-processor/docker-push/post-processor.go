package dockerpush

import (
	"fmt"
	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/docker-import"
	"github.com/mitchellh/packer/post-processor/docker-tag"
	"strings"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Login         bool
	LoginEmail    string `mapstructure:"login_email"`
	LoginUsername string `mapstructure:"login_username"`
	LoginPassword string `mapstructure:"login_password"`
	LoginServer   string `mapstructure:"login_server"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	Driver docker.Driver

	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := new(packer.MultiError)

	// Process templates
	templates := map[string]*string{
		"login_email":    &p.config.LoginEmail,
		"login_username": &p.config.LoginUsername,
		"login_password": &p.config.LoginPassword,
		"login_server":   &p.config.LoginServer,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != dockerimport.BuilderId &&
		artifact.BuilderId() != dockertag.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from docker-import and docker-tag artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	driver := p.Driver
	if driver == nil {
		// If no driver is set, then we use the real driver
		driver = &docker.DockerDriver{Tpl: p.config.tpl, Ui: ui}
	}

	if p.config.Login {
		ui.Message("Logging in...")
		err := driver.Login(
			p.config.LoginServer,
			p.config.LoginEmail,
			p.config.LoginUsername,
			p.config.LoginPassword)
		if err != nil {
			return nil, false, fmt.Errorf(
				"Error logging in to Docker: %s", err)
		}

		defer func() {
			ui.Message("Logging out...")
			if err := driver.Logout(p.config.LoginServer); err != nil {
				ui.Error(fmt.Sprintf("Error logging out: %s", err))
			}
		}()
	}

	// Get the name. We strip off any tags from the name because the
	// push doesn't use those.
	name := artifact.Id()

	if i := strings.Index(name, "/"); i >= 0 {
		// This should always be true because the / is required. But we have
		// to get the index to this so we don't accidentally strip off the port
		if j := strings.Index(name[i:], ":"); j >= 0 {
			name = name[:i+j]
		}
	}

	ui.Message("Pushing: " + name)
	if err := driver.Push(name); err != nil {
		return nil, false, err
	}

	return nil, false, nil
}
