//go:generate mapstructure-to-hcl2 -type Config

package dockerpush

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-docker/builder/docker"
	dockerimport "github.com/hashicorp/packer-plugin-docker/post-processor/docker-import"
	dockertag "github.com/hashicorp/packer-plugin-docker/post-processor/docker-tag"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const BuilderIdImport = "packer.post-processor.docker-import"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Login                  bool
	LoginUsername          string `mapstructure:"login_username"`
	LoginPassword          string `mapstructure:"login_password"`
	LoginServer            string `mapstructure:"login_server"`
	EcrLogin               bool   `mapstructure:"ecr_login"`
	docker.AwsAccessConfig `mapstructure:",squash"`

	ctx interpolate.Context
}

type PostProcessor struct {
	Driver docker.Driver

	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderIdImport,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.EcrLogin && p.config.LoginServer == "" {
		return fmt.Errorf("ECR login requires login server to be provided.")
	}
	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	if artifact.BuilderId() != dockerimport.BuilderId &&
		artifact.BuilderId() != dockertag.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from docker-import and docker-tag artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	driver := p.Driver
	if driver == nil {
		// If no driver is set, then we use the real driver
		driver = &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}
	}

	if p.config.EcrLogin {
		ui.Message("Fetching ECR credentials...")

		username, password, err := p.config.EcrGetLogin(p.config.LoginServer)
		if err != nil {
			return nil, false, false, err
		}

		p.config.LoginUsername = username
		p.config.LoginPassword = password
	}

	if p.config.Login || p.config.EcrLogin {
		ui.Message("Logging in...")
		err := driver.Login(
			p.config.LoginServer,
			p.config.LoginUsername,
			p.config.LoginPassword)
		if err != nil {
			return nil, false, false, fmt.Errorf(
				"Error logging in to Docker: %s", err)
		}

		defer func() {
			ui.Message("Logging out...")
			if err := driver.Logout(p.config.LoginServer); err != nil {
				ui.Error(fmt.Sprintf("Error logging out: %s", err))
			}
		}()
	}

	var tags []string
	switch t := artifact.State("docker_tags").(type) {
	case []string:
		tags = t
	case []interface{}:
		for _, name := range t {
			if n, ok := name.(string); ok {
				tags = append(tags, n)
			}
		}
	}

	names := []string{artifact.Id()}
	names = append(names, tags...)

	// Get the name.
	for _, name := range names {
		ui.Message("Pushing: " + name)
		if err := driver.Push(name); err != nil {
			return nil, false, false, err
		}
	}

	artifact = &docker.ImportArtifact{
		BuilderIdValue: BuilderIdImport,
		Driver:         driver,
		IdValue:        names[0],
		StateData:      map[string]interface{}{"docker_tags": tags},
	}

	return artifact, true, false, nil
}
