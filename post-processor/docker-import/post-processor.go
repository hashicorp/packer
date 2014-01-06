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

	Repository string `mapstructure:"repository"`
	Tag        string `mapstructure:"tag"`
	Dockerfile string `mapstructure:"dockerfile"`

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
	id := artifact.Id()
	ui.Say("Importing image: " + id)

	// TODO Set artifact ID so that docker-push can use it

	if p.config.Tag == "" {

		cmd := exec.Command("docker",
			"import",
			"-",
			p.config.Repository)

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

		if err := cmd.Start(); err != nil {
			ui.Say("Image import failed")
			return nil, false, err
		}

		go func() {
			io.Copy(stdin, file)
			// close stdin so that program will exit
			stdin.Close()
		}()

		cmd.Wait()

	} else {

		cmd := exec.Command("docker",
			"import",
			"-",
			p.config.Repository+":"+p.config.Tag)

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

		if err := cmd.Start(); err != nil {
			ui.Say("Image import failed")
			return nil, false, err
		}

		go func() {
			io.Copy(stdin, file)
			// close stdin so that program will exit
			stdin.Close()
		}()

		cmd.Wait()

	}

	// Process Dockerfile if provided
	if p.config.Dockerfile != "" {

		if p.config.Tag != "" {

			cmd := exec.Command("docker", "build", "-t="+p.config.Repository+":"+p.config.Tag, "-")

			stdin, err := cmd.StdinPipe()

			if err != nil {
				return nil, false, err
			}

			// open Dockerfile
			file, err := os.Open(p.config.Dockerfile)

			if err != nil {
				ui.Say("Could not open Dockerfile: " + p.config.Dockerfile)
				return nil, false, err
			}

			ui.Say(id)

			defer file.Close()

			if err := cmd.Start(); err != nil {
				ui.Say("Failed to build image: " + id)
				return nil, false, err
			}

			go func() {
				io.Copy(stdin, file)
				// close stdin so that program will exit
				stdin.Close()
			}()

			cmd.Wait()

		} else {

			cmd := exec.Command("docker", "build", "-t="+p.config.Repository, "-")

			stdin, err := cmd.StdinPipe()

			if err != nil {
				return nil, false, err
			}

			// open Dockerfile
			file, err := os.Open(p.config.Dockerfile)

			if err != nil {
				ui.Say("Could not open Dockerfile: " + p.config.Dockerfile)
				return nil, false, err
			}

			ui.Say(id)

			defer file.Close()

			if err := cmd.Start(); err != nil {
				ui.Say("Failed to build image: " + id)
				return nil, false, err
			}

			go func() {
				io.Copy(stdin, file)
				// close stdin so that program will exit
				stdin.Close()
			}()

			cmd.Wait()
		}

	}
	return nil, false, nil
}
