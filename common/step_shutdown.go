package common

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"log"
	"time"
)

type ShutdownConfig struct {
	Command    string `mapstructure:"shutdown_command"`
	RawTimeout string `mapstructure:"shutdown_timeout"`

	Timeout time.Duration
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
	Config *ShutdownConfig
}

func (s *StepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if s.Config.Command != "" {
		ui.Say("Executing shutdown command...")
		log.Printf("Shutdown command: %s", s.Config.Command)

		var stdout, stderr bytes.Buffer
		cmd := &packer.RemoteCmd{
			Command: s.Config.Command,
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

		err := vm.StartShutdown()
		if err != nil {
			state.Put("error", fmt.Errorf("Cannot shut down VM: %v", err))
			return multistep.ActionHalt
		}
	}

	log.Printf("Waiting max %s for shutdown to complete", s.Config.Timeout)
	err := vm.WaitForShutdown(ctx, s.Config.Timeout)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {}
