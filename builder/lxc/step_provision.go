package lxc

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// StepProvision provisions the instance within a chroot.
type StepProvision struct {}

func (s *StepProvision) Run(state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	config := state.Get("config").(*Config)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	// Create our communicator
	comm := &LxcAttachCommunicator{
		ContainerName: config.ContainerName,
		RootFs:        mountPath,
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
