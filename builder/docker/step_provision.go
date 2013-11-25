package docker

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
)

type StepProvision struct{}

func (s *StepProvision) Run(state multistep.StateBag) multistep.StepAction {
	containerId := state.Get("container_id").(string)
	tempDir := state.Get("temp_dir").(string)

	// Create the communicator that talks to Docker via various
	// os/exec tricks.
	comm := &Communicator{
		ContainerId:  containerId,
		HostDir:      tempDir,
		ContainerDir: "/packer-files",
	}

	prov := common.StepProvision{Comm: comm}
	return prov.Run(state)
}

func (s *StepProvision) Cleanup(state multistep.StateBag) {}
