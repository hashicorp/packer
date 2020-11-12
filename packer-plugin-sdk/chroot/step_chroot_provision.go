package chroot

import (
	"context"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/commonsteps"
)

// StepChrootProvision provisions the instance within a chroot.
type StepChrootProvision struct {
}

func (s *StepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	// Create our communicator
	comm := &Communicator{
		Chroot:     mountPath,
		CmdWrapper: wrappedCommand,
	}

	// Loads hook data from builder's state, if it has been set.
	hookData := commonsteps.PopulateProvisionHookData(state)

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

func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
