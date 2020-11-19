//go:generate mapstructure-to-hcl2 -type Config

package dockertag

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/docker"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	dockerimport "github.com/hashicorp/packer/post-processor/docker-import"
)

const BuilderId = "packer.post-processor.docker-tag"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Repository string `mapstructure:"repository"`
	// Kept for backwards compatability
	Tag   []string `mapstructure:"tag"`
	Tags  []string `mapstructure:"tags"`
	Force bool

	ctx interpolate.Context
}

type PostProcessor struct {
	Driver docker.Driver

	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// combine Tag and Tags fields
	allTags := p.config.Tags
	allTags = append(allTags, p.config.Tag...)

	p.config.Tags = allTags

	return nil

}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	if len(p.config.Tag) > 0 {
		ui.Say("Deprecation warning: \"tag\" option has been replaced with " +
			"\"tags\". In future versions of Packer, this configuration may " +
			"not work. Please call `packer fix` on your template to update.")
	}

	if artifact.BuilderId() != BuilderId &&
		artifact.BuilderId() != dockerimport.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only tag from Docker builder artifacts.",
			artifact.BuilderId())
		return nil, false, true, err
	}

	driver := p.Driver
	if driver == nil {
		// If no driver is set, then we use the real driver
		driver = &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}
	}

	importRepo := p.config.Repository
	var lastTaggedRepo = importRepo
	RepoTags := []string{}

	if len(p.config.Tags) > 0 {
		for _, tag := range p.config.Tags {
			local := importRepo + ":" + tag
			ui.Message("Tagging image: " + artifact.Id())
			ui.Message("Repository: " + local)

			err := driver.TagImage(artifact.Id(), local, p.config.Force)
			if err != nil {
				return nil, false, true, err
			}

			RepoTags = append(RepoTags, local)
			lastTaggedRepo = local
		}
	} else {
		ui.Message("Tagging image: " + artifact.Id())
		ui.Message("Repository: " + importRepo)
		err := driver.TagImage(artifact.Id(), importRepo, p.config.Force)
		if err != nil {
			return nil, false, true, err
		}
	}

	// Build the artifact
	artifact = &docker.ImportArtifact{
		BuilderIdValue: BuilderId,
		Driver:         driver,
		IdValue:        lastTaggedRepo,
		StateData:      map[string]interface{}{"docker_tags": RepoTags},
	}

	// If we tag an image and then delete it, there was no point in creating the
	// tag. Override users to force us to always keep the input artifact.
	return artifact, true, true, nil
}
