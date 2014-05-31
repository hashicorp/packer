package dockerimport

import (
	"fmt"
	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"strings"
)

const BuilderId = "packer.post-processor.docker-import"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Repository      string   `mapstructure:"repository"`
	Tag             string   `mapstructure:"tag"`
	DockerfileLines []string `mapstructure:"dockerfile_lines"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
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
		"repository": &p.config.Repository,
		"tag":        &p.config.Tag,
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
	if artifact.BuilderId() != docker.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from Docker builder artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	importRepo := p.config.Repository
	if p.config.Tag != "" {
		importRepo += ":" + p.config.Tag
	}

	driver := &docker.DockerDriver{Tpl: p.config.tpl, Ui: ui}

	ui.Message("Importing image: " + artifact.Id())
	ui.Message("Repository: " + importRepo)
	id, err := driver.Import(artifact.Files()[0], importRepo)
	if err != nil {
		return nil, false, err
	}

	ui.Message("Imported ID: " + id)

	if len(p.config.DockerfileLines) > 0 {
		// Use Dockerfile
		dockerfile := "FROM " + importRepo + "\n" +
			strings.Join(p.config.DockerfileLines, "\n")

		ui.Message("Updating with Dockerfile:")
		ui.Message(dockerfile)
		err := driver.BuildFromStdin(dockerfile, importRepo)

		if err != nil {
			return nil, false, err
		}
	}

	// Build the artifact
	artifact = &docker.ImportArtifact{
		BuilderIdValue: BuilderId,
		Driver:         driver,
		IdValue:        importRepo,
	}

	return artifact, false, nil
}
