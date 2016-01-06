package docker

import (
	"github.com/mitchellh/multistep"
)

type StepConnectDocker struct{}

func (s *StepConnectDocker) Run(state multistep.StateBag) multistep.StepAction {
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

	// Create the communicator that talks to Docker via various
	// os/exec tricks.
	comm := &Communicator{
		ContainerId:  containerId,
		HostDir:      tempDir,
		ContainerDir: "/packer-files",
		Version:      version,
		Config:       config,
	}

	state.Put("communicator", comm)
	return multistep.ActionContinue
}

func (s *StepConnectDocker) Cleanup(state multistep.StateBag) {}
