package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
	"fmt"
	"log"
	"time"
	"bytes"
	"errors"
)

type ShutdownConfig struct {
	Command    string `mapstructure:"shutdown_command"`
	RawTimeout string `mapstructure:"shutdown_timeout"`
	Timeout    time.Duration
}

func (c *ShutdownConfig) Prepare() []error {
	var errs []error

	if c.RawTimeout != "" {
		timeout, err := time.ParseDuration(c.RawTimeout)
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
			return errs
		}
		c.Timeout = timeout
	} else {
		c.Timeout = 5 * time.Minute
	}

	return nil
}

type StepShutdown struct {
	config *ShutdownConfig
}

func (s *StepShutdown) Run(state multistep.StateBag) multistep.StepAction {
	// is set during the communicator.StepConnect
	comm := state.Get("communicator").(packer.Communicator)
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*object.VirtualMachine)
	d := state.Get("driver").(Driver)

	ui.Say("Shut down VM...")

	if s.config.Command != "" {
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", s.config.Command)

		var stdout, stderr bytes.Buffer
		cmd := &packer.RemoteCmd{
			Command: s.config.Command,
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		if err := comm.Start(cmd); err != nil {
			state.Put("error", fmt.Errorf("Failed to send shutdown command: %s", err))
			return multistep.ActionHalt
		}
	} else {
		ui.Say("Forcibly halting virtual machine...")

		err := vm.ShutdownGuest(d.ctx)
		if err != nil {
			state.Put("error", fmt.Errorf("Cannot shut down VM: %v", err))
			return multistep.ActionHalt
		}
	}

	// Wait for the machine to actually shut down
	log.Printf("Waiting max %s for shutdown to complete", s.config.Timeout)
	shutdownTimer := time.After(s.config.Timeout)
	for {
		powerState, err := vm.PowerState(d.ctx)
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
