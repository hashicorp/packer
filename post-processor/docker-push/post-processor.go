package dockerpush

import (
	"fmt"

	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/docker-import"
	"github.com/mitchellh/packer/post-processor/docker-tag"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Login         bool
	LoginEmail    string `mapstructure:"login_email"`
	LoginUsername string `mapstructure:"login_username"`
	LoginPassword string `mapstructure:"login_password"`
	LoginServer   string `mapstructure:"login_server"`

	ctx interpolate.Context
}

type PostProcessor struct {
	Driver docker.Driver

	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
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
		driver = &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}
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

	// Get the name.
	name := artifact.Id()

	ui.Message("Pushing: " + name)
	if err := driver.Push(name); err != nil {
		return nil, false, err
	}

	return nil, false, nil
}
