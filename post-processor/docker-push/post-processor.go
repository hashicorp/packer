package dockerpush

import (
	"fmt"
	"os"

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

	Login                   bool
	LoginEmail              string `mapstructure:"login_email"`
	LoginUsername           string `mapstructure:"login_username"`
	LoginPassword           string `mapstructure:"login_password"`
	LoginServer             string `mapstructure:"login_server"`
	EcrLogin                bool   `mapstructure:"ecr_login"`
	docker.DockerHostConfig `mapstructure:",squash"`
	docker.AwsAccessConfig  `mapstructure:",squash"`

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

	var errs *packer.MultiError

	if es := p.config.DockerHostConfig.Prepare(); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if p.config.EcrLogin && p.config.LoginServer == "" {
		errs = packer.MultiErrorAppend(fmt.Errorf("ECR login requires login server to be provided."))
	}
	if errs != nil && len(errs.Errors) > 0 {
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
		if os.Getenv("PACKER_DOCKER_API") != "" {
			driver = &docker.DockerApiDriver{Ctx: &p.config.ctx, Config: p.config.DockerHostConfig, Ui: ui}
		} else {
			driver = &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}
		}

		err := driver.Verify()
		if err != nil {
			return nil, false, err
		}
	}

	if p.config.EcrLogin {
		ui.Message("Fetching ECR credentials...")

		username, password, err := p.config.EcrGetLogin(p.config.LoginServer)
		if err != nil {
			return nil, false, err
		}

		p.config.LoginUsername = username
		p.config.LoginPassword = password
	}

	if p.config.Login || p.config.EcrLogin {
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
