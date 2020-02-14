package lxd

import (
	"context"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepProvision provisions the container
type StepProvision struct{}

func (s *StepProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	// Create our communicator
	comm := &Communicator{
		ContainerName: config.ContainerName,
		CmdWrapper:    wrappedCommand,
	}

	// Loads hook data from builder's state, if it has been set.
	hookData := common.PopulateProvisionHookData(state)

	// Update state generated_data with complete hookData
	// to make them accessible by post-processors
	state.Put("generated_data", hookData)

	// Provision
	log.Println("Running the provision hook")
	if err := hook.Run(ctx, packer.HookProvision, ui, comm, hookData); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepProvision) Cleanup(state multistep.StateBag) {}
