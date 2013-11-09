package docker

import (
	"github.com/mitchellh/packer/packer"
	"os/exec"
)

type DockerDriver struct {
	Ui packer.Ui
}

func (d *DockerDriver) Pull(image string) error {
	cmd := exec.Command("docker", "pull", image)
	return runAndStream(cmd, d.Ui)
}
