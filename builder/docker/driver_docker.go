package docker

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type DockerDriver struct {
	Ui  packer.Ui
	Tpl *packer.ConfigTemplate
}

func (d *DockerDriver) DeleteImage(id string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("docker", "rmi", id)
	cmd.Stderr = &stderr

	log.Printf("Deleting image: %s", id)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error deleting image: %s\nStderr: %s",
			err, stderr.String())
		return err
	}

	return nil
}

func (d *DockerDriver) Export(id string, dst io.Writer) error {
	var stderr bytes.Buffer
	cmd := exec.Command("docker", "export", id)
	cmd.Stdout = dst
	cmd.Stderr = &stderr

	log.Printf("Exporting container: %s", id)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error exporting: %s\nStderr: %s",
			err, stderr.String())
		return err
	}

	return nil
}

func (d *DockerDriver) Import(path string, repo string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("docker", "import", "-", repo)
	cmd.Stdout = &stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	// There should be only one artifact of the Docker builder
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, file)
	}()

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error importing container: %s", err)
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *DockerDriver) Pull(image string) error {
	cmd := exec.Command("docker", "pull", image)
	return runAndStream(cmd, d.Ui)
}

func (d *DockerDriver) Push(name string) error {
	cmd := exec.Command("docker", "push", name)
	return runAndStream(cmd, d.Ui)
}

func (d *DockerDriver) StartContainer(config *ContainerConfig) (string, error) {
	// Build up the template data
	var tplData startContainerTemplate
	tplData.Image = config.Image
	if len(config.Volumes) > 0 {
		volumes := make([]string, 0, len(config.Volumes))
		for host, guest := range config.Volumes {
			volumes = append(volumes, fmt.Sprintf("%s:%s", host, guest))
		}

		tplData.Volumes = strings.Join(volumes, ",")
	}

	// Args that we're going to pass to Docker
	args := config.RunCommand
	for i, v := range args {
		var err error
		args[i], err = d.Tpl.Process(v, &tplData)
		if err != nil {
			return "", err
		}
	}
	d.Ui.Message(fmt.Sprintf(
		"Run command: docker %s", strings.Join(args, " ")))

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
		if _, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("Docker exited with a non-zero exit status.\nStderr: %s",
				stderr.String())
		}

		return "", err
	}

	// Capture the container ID, which is alone on stdout
	return strings.TrimSpace(stdout.String()), nil
}

func (d *DockerDriver) StopContainer(id string) error {
	return exec.Command("docker", "kill", id).Run()
}

func (d *DockerDriver) Verify() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return err
	}

	return nil
}
