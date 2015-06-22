package dockersave

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

const BuilderId = "packer.post-processor.docker-save"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Path string `mapstructure:"path"`

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
			"Unknown artifact type: %s\nCan only save Docker builder artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	path := p.config.Path

	// Open the file that we're going to write to
	f, err := os.Create(path)
	if err != nil {
		err := fmt.Errorf("Error creating output file: %s", err)
		return nil, false, err
	}

	driver := p.Driver
	if driver == nil {
		// If no driver is set, then we use the real driver
		driver = &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}
	}

	ui.Message("Saving image: " + artifact.Id())

	if err := driver.SaveImage(artifact.Id(), f); err != nil {
		f.Close()
		os.Remove(f.Name())

		return nil, false, err
	}

	f.Close()
	ui.Message("Saved to: " + path)

	return artifact, true, nil
}
