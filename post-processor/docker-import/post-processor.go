package dockerimport

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"io"
	"os"
	"os/exec"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Dockerfile string `mapstructure:"dockerfile"`
	Repository string `mapstructure:"repository"`
	Tag        string `mapstructure:"tag"`

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
		"dockerfile": &p.config.Dockerfile,
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
	importRepo := p.config.Repository
	if p.config.Tag != "" {
		importRepo += ":" + p.config.Tag
	}

	cmd := exec.Command("docker", "import", "-", importRepo)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, false, err
	}

	// There should be only one artifact of the Docker builder
	file, err := os.Open(artifact.Files()[0])
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	ui.Message("Importing image: " + artifact.Id())
	ui.Message("Repository: " + importRepo)

	if err := cmd.Start(); err != nil {
		return nil, false, err
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, file)
	}()

	cmd.Wait()

	// Process Dockerfile if provided
	if p.config.Dockerfile != "" {
		cmd := exec.Command("docker", "build", "-t="+importRepo, "-")

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, false, err
		}

		// open Dockerfile
		file, err := os.Open(p.config.Dockerfile)
		if err != nil {
			err = fmt.Errorf("Couldn't open Dockerfile: %s", err)
			return nil, false, err
		}
		defer file.Close()

		ui.Message("Running docker build with Dockerfile: " + p.config.Dockerfile)
		if err := cmd.Start(); err != nil {
			err = fmt.Errorf("Failed to start docker build: %s", err)
			return nil, false, err
		}

		go func() {
			defer stdin.Close()
			io.Copy(stdin, file)
		}()

		cmd.Wait()

	}
	return nil, false, nil
}
