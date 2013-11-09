package docker

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepProvision struct{}

func (s *StepProvision) Run(state multistep.StateBag) multistep.StepAction {
	containerId := state.Get("container_id").(string)
	hook := state.Get("hook").(packer.Hook)
	tempDir := state.Get("temp_dir").(string)
	ui := state.Get("ui").(packer.Ui)

	// Create the communicator that talks to Docker via various
	// os/exec tricks.
	comm := &Communicator{
		ContainerId:  containerId,
		HostDir:      tempDir,
		ContainerDir: "/packer-files",
	}

	// Run the provisioning hook
	if err := hook.Run(packer.HookProvision, ui, comm, nil); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepProvision) Cleanup(state multistep.StateBag) {}
