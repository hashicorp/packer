package docker

import (
	"bytes"
	"errors"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"os/exec"
)

type Config struct {
	Registry string
	Username string
	Password string
	Email    string
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raw ...interface{}) error {
	if err := mapstructure.Decode(raw, &p.config); err != nil {
		return err
	}

	if p.config.Registry == "" {
		p.config.Registry = "registry.docker.io"
	}

	if p.config.Username == "" {
		return errors.New("Username is required to push docker image")
	}

	if p.config.Password == "" {
		return errors.New("Password is required to push docker image")
	}

	return nil
}
func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	id := artifact.Id()
	ui.Say("Pushing imgage: " + id)

	// TODO: docker login

	stdout := new(bytes.Buffer)
	cmd := exec.Command("docker", "push", id)
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		ui.Say("Failed to push image: " + id)
		return nil, true, err
	}

	return nil, true, nil
}
