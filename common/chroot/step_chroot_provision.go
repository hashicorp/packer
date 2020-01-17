package chroot

import (
	"context"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepChrootProvision provisions the instance within a chroot.
type StepChrootProvision struct {
}

func (s *StepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	// At this point variables are valid but not assigned a value
	// Retrieve generated data from builders and assign default variables. Like in step_provision

	// Create our communicator
	comm := &Communicator{
		Chroot:     mountPath,
		CmdWrapper: wrappedCommand,
	}

	// Provision
	log.Println("Running the provision hook")
	if err := hook.Run(ctx, packer.HookProvision, ui, comm, nil); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
