package docker

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
)

type StepProvision struct{}

func (s *StepProvision) Run(state multistep.StateBag) multistep.StepAction {
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
	}

	prov := common.StepProvision{Comm: comm}
	return prov.Run(state)
}

func (s *StepProvision) Cleanup(state multistep.StateBag) {}
