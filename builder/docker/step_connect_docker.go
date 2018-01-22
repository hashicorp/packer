package docker

import (
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"os/exec"
	"strings"

	"github.com/mitchellh/multistep"
)

type StepConnectDocker struct{}

func (s *StepConnectDocker) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	containerId := state.Get("container_id").(string)
	driver := state.Get("driver").(Driver)
	tempDir := state.Get("temp_dir").(string)

	// Get the version so we can pass it to the communicator
	version, err := driver.Version()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	containerUser, err := getContainerUser(containerId)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Create the communicator that talks to Docker via various
	// os/exec tricks.
	comm := &Communicator{
		ContainerID:   containerId,
		HostDir:       tempDir,
		ContainerDir:  config.ContainerDir,
		Version:       version,
		Config:        config,
		ContainerUser: containerUser,
	}

	state.Put("communicator", comm)
	return multistep.ActionContinue
}

func (s *StepConnectDocker) Cleanup(state multistep.StateBag) {}

func getContainerUser(containerId string) (string, error) {
	inspectArgs := []string{"docker", "inspect", "--format", "{{.Config.User}}", containerId}
	stdout, err := exec.Command(inspectArgs[0], inspectArgs[1:]...).Output()
	if err != nil {
		errStr := fmt.Sprintf("Failed to inspect the container: %s", err)
		if ee, ok := err.(*exec.ExitError); ok {
			errStr = fmt.Sprintf("%s, %s", errStr, ee.Stderr)
		}
		return "", fmt.Errorf(errStr)
	}
	return strings.TrimSpace(string(stdout)), nil
}
