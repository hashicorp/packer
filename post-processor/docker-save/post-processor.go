package dockersave

import (
	"fmt"
	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/docker-import"
	"os"
)

const BuilderId = "packer.post-processor.docker-save"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Path string `mapstructure:"path"`

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

	templates := map[string]*string{
		"path": &p.config.Path,
	}

	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}

		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", key, err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil

}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != dockerimport.BuilderId {
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
		driver = &docker.DockerDriver{Tpl: p.config.tpl, Ui: ui}
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
