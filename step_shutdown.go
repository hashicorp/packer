package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
	"fmt"
	"log"
	"time"
	"bytes"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
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
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	d := state.Get("driver").(*driver.Driver)
	vm := state.Get("vm").(*object.VirtualMachine)

	if s.config.Command != "" {
		ui.Say("Executing shutdown command...")
		log.Printf("Shutdown command: %s", s.config.Command)

		var stdout, stderr bytes.Buffer
		cmd := &packer.RemoteCmd{
			Command: s.config.Command,
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		err := comm.Start(cmd)
		if err != nil {
			state.Put("error", fmt.Errorf("Failed to send shutdown command: %s", err))
			return multistep.ActionHalt
		}
	} else {
		ui.Say("Shut down VM...")

		err := d.StartShutdown(vm)
		if err != nil {
			state.Put("error", fmt.Errorf("Cannot shut down VM: %v", err))
			return multistep.ActionHalt
		}
	}

	log.Printf("Waiting max %s for shutdown to complete", s.config.Timeout)
	err := d.WaitForShutdown(vm, s.config.Timeout)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("VM stopped")
	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {}
