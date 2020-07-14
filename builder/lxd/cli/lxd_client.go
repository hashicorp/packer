package cli_client

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/hashicorp/packer/packer"
)

// LXDClient connect to the LXD server via the `lxc` command line interface.
type LXDClient struct{}

func (s *LXDClient) DeleteImage(name string) error {
	_, err := lxdCommand("image", "delete", name)
	return err
}

func (s *LXDClient) LaunchContainer(name string, image string, profile string, launchConfig map[string]string) error {
	launch_args := []string{
		"launch", "--ephemeral=false", fmt.Sprintf("--profile=%s", profile), image, name,
	}
	for k, v := range launchConfig {
		launch_args = append(launch_args, "--config", fmt.Sprintf("%s=%s", k, v))
	}
	_, err := lxdCommand(launch_args...)
	return err
}

func (s *LXDClient) PublishContainer(name string, outImage string, publishProperties map[string]string) (string, error) {
	publish_args := []string{
		"publish", name, "--alias", outImage,
	}

	for k, v := range publishProperties {
		publish_args = append(publish_args, fmt.Sprintf("%s=%s", k, v))
	}

	stdoutString, err := lxdCommand(publish_args...)
	r := regexp.MustCompile("([0-9a-fA-F]+)$")
	fingerprint := r.FindAllStringSubmatch(stdoutString, -1)[0][0]
	return fingerprint, err
}

func (s *LXDClient) StopContainer(name string) error {
	stop_args := []string{
		// We created the container with "--ephemeral=false" so we know it is safe to stop.
		"stop", name,
	}
	_, err := lxdCommand(stop_args...)
	return err
}

func (s *LXDClient) DeleteContainer(name string) error {
	cleanup_args := []string{
		"delete", "--force", name,
	}
	_, err := lxdCommand(cleanup_args...)
	return err
}

func (s *LXDClient) ExecuteContainer(name string, wrapper func(string) (string, error), cmd *packer.RemoteCmd) error {
	log.Printf("Executing with lxc exec in container: %s %s", name, cmd.Command)
	command, err := wrapper(
		fmt.Sprintf("lxc exec %s -- /bin/sh -c \"%s\"", name, cmd.Command))
	if err != nil {
		return  err
	}

	localCmd := exec.Command("/bin/sh", "-c", command)
	log.Printf("Executing lxc exec: %s %#v", localCmd.Path, localCmd.Args)

	if err != nil {
		return err
	}

	localCmd.Stdin = cmd.Stdin
	localCmd.Stdout = cmd.Stdout
	localCmd.Stderr = cmd.Stderr
	if err := localCmd.Start(); err != nil {
		return  err
	}	

	go func() {
		exitStatus := 0
		if err := localCmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitStatus = 1

				// There is no process-independent way to get the REAL
				// exit status so we just try to go deeper.
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitStatus = status.ExitStatus()
				}
			}
		}

		log.Printf(
			"lxc exec execution exited with '%d': '%s'",
			exitStatus, cmd.Command)
		cmd.SetExited(exitStatus)
	}()

	return nil
}

// Yeah...LXD calls `lxc` because the command line is different between the
// packages. This should also avoid a naming collision between the LXC builder.
func lxdCommand(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing lxc command: %#v", args)
	cmd := exec.Command("lxc", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("LXD command error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return stdoutString, err
}

// ShellCommand takes a command string and returns an *exec.Cmd to execute
// it within the context of a shell (/bin/sh).
func ShellCommand(command string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", command)
}
