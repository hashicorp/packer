package lxc

import (
	"context"
	"log"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepProvision provisions the instance within a chroot.
type StepProvision struct{}

func (s *StepProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	config := state.Get("config").(*Config)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packersdk.Ui)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	// Create our communicator
	comm := &LxcAttachCommunicator{
		ContainerName: config.ContainerName,
		AttachOptions: config.AttachOptions,
		RootFs:        mountPath,
		CmdWrapper:    wrappedCommand,
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

func (s *StepProvision) Cleanup(state multistep.StateBag) {}
