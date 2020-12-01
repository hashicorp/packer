//go:generate mapstructure-to-hcl2 -type Config

package dockerimport

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/docker"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/post-processor/artifice"
)

const BuilderId = "packer.post-processor.docker-import"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Repository string   `mapstructure:"repository"`
	Tag        string   `mapstructure:"tag"`
	Changes    []string `mapstructure:"changes"`

	ctx interpolate.Context
}

type PostProcessor struct {
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

	return nil

}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	switch artifact.BuilderId() {
	case docker.BuilderId, artifice.BuilderId:
		break
	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from Docker builder "+
				"and Artifice post-processor artifacts. If you are getting this "+
				"error after having run the docker builder, it may be because you "+
				"set commit: true in your Docker builder, so the image is "+
				"already imported. ",
			artifact.BuilderId())
		return nil, false, false, err
	}

	importRepo := p.config.Repository
	if p.config.Tag != "" {
		importRepo += ":" + p.config.Tag
	}

	driver := &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}

	ui.Message("Importing image: " + artifact.Id())
	ui.Message("Repository: " + importRepo)
	id, err := driver.Import(artifact.Files()[0], p.config.Changes, importRepo)
	if err != nil {
		return nil, false, false, err
	}

	ui.Message("Imported ID: " + id)

	// Build the artifact
	artifact = &docker.ImportArtifact{
		BuilderIdValue: BuilderId,
		Driver:         driver,
		IdValue:        importRepo,
	}

	return artifact, false, false, nil
}
