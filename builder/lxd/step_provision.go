package lxd

import (
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"log"
)

// StepProvision provisions the container
type StepProvision struct{}

func (s *StepProvision) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	// Create our communicator
	comm := &Communicator{
		ContainerName: config.ContainerName,
		CmdWrapper:    wrappedCommand,
	}

	// Provision
	log.Println("Running the provision hook")
	if err := hook.Run(packer.HookProvision, ui, comm, nil); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepProvision) Cleanup(state multistep.StateBag) {}
