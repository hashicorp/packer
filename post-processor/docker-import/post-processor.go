package dockerimport

import (
	"fmt"

	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/artifice"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "packer.post-processor.docker-import"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Repository string `mapstructure:"repository"`
	Tag        string `mapstructure:"tag"`

	ctx interpolate.Context
}

type PostProcessor struct {
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
	switch artifact.BuilderId() {
	case docker.BuilderId, artifice.BuilderId:
		break
	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from Docker builder and Artifice post-processor artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	importRepo := p.config.Repository
	if p.config.Tag != "" {
		importRepo += ":" + p.config.Tag
	}

	driver := &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}

	ui.Message("Importing image: " + artifact.Id())
	ui.Message("Repository: " + importRepo)
	id, err := driver.Import(artifact.Files()[0], importRepo)
	if err != nil {
		return nil, false, err
	}

	ui.Message("Imported ID: " + id)

	// Build the artifact
	artifact = &docker.ImportArtifact{
		BuilderIdValue: BuilderId,
		Driver:         driver,
		IdValue:        importRepo,
	}

	return artifact, false, nil
}
