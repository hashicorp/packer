package chroot

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
)

// StepChrootProvision provisions the instance within a chroot.
type StepChrootProvision struct {
	mounts []string
}

type WrappedCommandTemplate struct {
	Command string
}

func (s *StepChrootProvision) Run(state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	mountPath := state.Get("mount_path").(string)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	chrootCmd := func(command string) *exec.Cmd {
		return ChrootCommand(mountPath, command)
	}
	wrappedCommand := func(command string) *exec.Cmd {
		wrapped, err := config.tpl.Process(config.CommandWrapper, &WrappedCommandTemplate{
			Command: command,
		})
		if err != nil {
			ui.Error(err.Error())
		}
		return ShellCommand(wrapped)
	}

	state.Put("chrootCmd", chrootCmd)
	state.Put("wrappedCommand", wrappedCommand)

	// Create our communicator
	comm := &Communicator{
		Chroot:         mountPath,
		ChrootCmd:      chrootCmd,
		wrappedCommand: wrappedCommand,
	}

	// Provision
	log.Println("Running the provision hook")
	if err := hook.Run(packer.HookProvision, ui, comm, nil); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
