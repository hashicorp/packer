package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
	"context"
	"fmt"
	"log"
	"time"
	"bytes"
	"errors"
)

type StepShutdown struct{
	Command    string
	ShutdownTimeout time.Duration
}

func (s *StepShutdown) Run(state multistep.StateBag) multistep.StepAction {
	// is set during the communicator.StepConnect
	comm := state.Get("communicator").(packer.Communicator)
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*object.VirtualMachine)
	ctx := state.Get("ctx").(context.Context)

	ui.Say("VM shutdown...")

	if s.Command != "" {
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", s.Command)

		var stdout, stderr bytes.Buffer
		cmd := &packer.RemoteCmd{
			Command: s.Command,
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		if err := comm.Start(cmd); err != nil {
			err := fmt.Errorf("Failed to send shutdown command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		ui.Say("Forcibly halting virtual machine...")

		err := vm.ShutdownGuest(ctx)
		if err != nil {
			state.Put("error", fmt.Errorf("Could not shutdown guest: %v", err))
			return multistep.ActionHalt
		}
	}

	// Wait for the machine to actually shut down
	log.Printf("Waiting max %s for shutdown to complete", s.ShutdownTimeout)
	shutdownTimer := time.After(s.ShutdownTimeout)
	for {
		powerState, err := vm.PowerState(ctx)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if powerState == "poweredOff" {
			break
		}

		select {
		case <-shutdownTimer:
			err := errors.New("Timeout while waiting for machine to shut down.")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		default:
			time.Sleep(150 * time.Millisecond)
		}
	}

	ui.Say("VM stopped")
	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {}

