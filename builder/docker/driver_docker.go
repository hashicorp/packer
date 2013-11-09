package docker

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
	"strings"
)

type DockerDriver struct {
	Ui packer.Ui
}

func (d *DockerDriver) Pull(image string) error {
	cmd := exec.Command("docker", "pull", image)
	return runAndStream(cmd, d.Ui)
}

func (d *DockerDriver) StartContainer(config *ContainerConfig) (string, error) {
	// Args that we're going to pass to Docker
	args := []string{"run", "-d", "-i", "-t"}

	if len(config.Volumes) > 0 {
		volumes := make([]string, 0, len(config.Volumes))
		for host, guest := range config.Volumes {
			volumes = append(volumes, fmt.Sprintf("%s:%s", host, guest))
		}

		args = append(args, "-v", strings.Join(volumes, ","))
	}

	args = append(args, config.Image, "/bin/bash")

	// Start the container
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Starting container with args: %v", args)
	if err := cmd.Start(); err != nil {
		return "", err
	}

	log.Println("Waiting for container to finish starting")
	if err := cmd.Wait(); err != nil {
		return "", err
	}

	// Capture the container ID, which is alone on stdout
	return strings.TrimSpace(stdout.String()), nil
}

func (d *DockerDriver) StopContainer(id string) error {
	return exec.Command("docker", "kill", id).Run()
}
