package docker

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type DockerDriver struct {
	Ui  packersdk.Ui
	Ctx *interpolate.Context

	l sync.Mutex
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

func (d *DockerDriver) Commit(id string, author string, changes []string, message string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	args := []string{"commit"}
	if author != "" {
		args = append(args, "--author", author)
	}
	for _, change := range changes {
		args = append(args, "--change", change)
	}
	if message != "" {
		args = append(args, "--message", message)
	}
	args = append(args, id)

	log.Printf("Committing container with args: %v", args)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error committing container: %s\nStderr: %s",
			err, stderr.String())
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
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

func (d *DockerDriver) Import(path string, changes []string, repo string) (string, error) {
	var stdout, stderr bytes.Buffer

	args := []string{"import"}

	for _, change := range changes {
		args = append(args, "--change", change)
	}

	args = append(args, "-")
	args = append(args, repo)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
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

	log.Printf("Importing tarball with args: %v", args)

	if err := cmd.Start(); err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, file)
	}()

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("Error importing container: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *DockerDriver) IPAddress(id string) (string, error) {
	var stderr, stdout bytes.Buffer
	cmd := exec.Command(
		"docker",
		"inspect",
		"--format",
		"{{ .NetworkSettings.IPAddress }}",
		id)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *DockerDriver) Sha256(id string) (string, error) {
	var stderr, stdout bytes.Buffer
	cmd := exec.Command(
		"docker",
		"inspect",
		"--format",
		"{{ .Id }}",
		id)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: %s\n\nStderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (d *DockerDriver) Login(repo, user, pass string) error {
	d.l.Lock()

	version_running, err := d.Version()
	if err != nil {
		d.l.Unlock()
		return err
	}

	// Version 17.07.0 of Docker adds support for the new
	// `--password-stdin` option which can be used to offer
	// password via the standard input, rather than passing
	// the password and/or token using a command line switch.
	constraint, err := version.NewConstraint(">= 17.07.0")
	if err != nil {
		d.l.Unlock()
		return err
	}

	cmd := exec.Command("docker")
	cmd.Args = append(cmd.Args, "login")

	if user != "" {
		cmd.Args = append(cmd.Args, "-u", user)
	}

	if pass != "" {
		if constraint.Check(version_running) {
			cmd.Args = append(cmd.Args, "--password-stdin")

			stdin, err := cmd.StdinPipe()
			if err != nil {
				d.l.Unlock()
				return err
			}
			io.WriteString(stdin, pass)
			stdin.Close()
		} else {
			cmd.Args = append(cmd.Args, "-p", pass)
		}
	}

	if repo != "" {
		cmd.Args = append(cmd.Args, repo)
	}

	err = runAndStream(cmd, d.Ui)
	if err != nil {
		d.l.Unlock()
		return err
	}

	return nil
}

func (d *DockerDriver) Logout(repo string) error {
	args := []string{"logout"}
	if repo != "" {
		args = append(args, repo)
	}

	cmd := exec.Command("docker", args...)
	err := runAndStream(cmd, d.Ui)
	d.l.Unlock()
	return err
}

func (d *DockerDriver) Pull(image string) error {
	cmd := exec.Command("docker", "pull", image)
	return runAndStream(cmd, d.Ui)
}

func (d *DockerDriver) Push(name string) error {
	cmd := exec.Command("docker", "push", name)
	return runAndStream(cmd, d.Ui)
}

func (d *DockerDriver) SaveImage(id string, dst io.Writer) error {
	var stderr bytes.Buffer
	cmd := exec.Command("docker", "save", id)
	cmd.Stdout = dst
	cmd.Stderr = &stderr

	log.Printf("Exporting image: %s", id)
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

func (d *DockerDriver) StartContainer(config *ContainerConfig) (string, error) {
	// Build up the template data
	var tplData startContainerTemplate
	tplData.Image = config.Image
	ictx := *d.Ctx
	ictx.Data = &tplData

	// Args that we're going to pass to Docker
	args := []string{"run"}
	for _, v := range config.Device {
		args = append(args, "--device", v)
	}
	for _, v := range config.CapAdd {
		args = append(args, "--cap-add", v)
	}
	for _, v := range config.CapDrop {
		args = append(args, "--cap-drop", v)
	}
	if config.Privileged {
		args = append(args, "--privileged")
	}
	for _, v := range config.TmpFs {
		args = append(args, "--tmpfs", v)
	}
	for host, guest := range config.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", host, guest))
	}
	for _, v := range config.RunCommand {
		v, err := interpolate.Render(v, &ictx)
		if err != nil {
			return "", err
		}

		args = append(args, v)
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
	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		return err
	}
	return nil
}

func (d *DockerDriver) KillContainer(id string) error {
	if err := exec.Command("docker", "kill", id).Run(); err != nil {
		return err
	}

	return exec.Command("docker", "rm", id).Run()
}

func (d *DockerDriver) TagImage(id string, repo string, force bool) error {
	args := []string{"tag"}

	// detect running docker version before tagging
	// flag `force` for docker tagging was removed after Docker 1.12.0
	// to keep its backward compatibility, we are not going to remove `force`
	// option, but to ignore it when Docker version >= 1.12.0
	//
	// for more detail, please refer to the following links:
	// - https://docs.docker.com/engine/deprecated/#/f-flag-on-docker-tag
	// - https://github.com/docker/docker/pull/23090
	version_running, err := d.Version()
	if err != nil {
		return err
	}

	version_deprecated, err := version.NewVersion("1.12.0")
	if err != nil {
		// should never reach this line
		return err
	}

	if force {
		if version_running.LessThan(version_deprecated) {
			args = append(args, "-f")
		} else {
			// do nothing if Docker version >= 1.12.0
			log.Printf("[WARN] option: \"force\" will be ignored here")
			log.Printf("since it was removed after Docker 1.12.0 released")
		}
	}
	args = append(args, id, repo)

	var stderr bytes.Buffer
	cmd := exec.Command("docker", args...)
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("Error tagging image: %s\nStderr: %s",
			err, stderr.String())
		return err
	}

	return nil
}

func (d *DockerDriver) Verify() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return err
	}

	return nil
}

func (d *DockerDriver) Version() (*version.Version, error) {
	output, err := exec.Command("docker", "-v").Output()
	if err != nil {
		return nil, err
	}

	match := regexp.MustCompile(version.VersionRegexpRaw).FindSubmatch(output)
	if match == nil {
		return nil, fmt.Errorf("unknown version: %s", output)
	}

	return version.NewVersion(string(match[0]))
}
