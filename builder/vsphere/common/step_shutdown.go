//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ShutdownConfig

package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type ShutdownConfig struct {
	// Specify a VM guest shutdown command. VMware guest tools are used by
	// default.
	Command string `mapstructure:"shutdown_command"`
	// Amount of time to wait for graceful VM shutdown.
	// Defaults to 5m or five minutes.
	Timeout time.Duration `mapstructure:"shutdown_timeout"`
	// Packer normally halts the virtual machine after all provisioners have
	// run when no `shutdown_command` is defined. If this is set to `true`, Packer
	// *will not* halt the virtual machine but will assume that you will send the stop
	// signal yourself through the preseed.cfg or your final provisioner.
	// Packer will wait for a default of five minutes until the virtual machine is shutdown.
	// The timeout can be changed using `shutdown_timeout` option.
	DisableShutdown bool `mapstructure:"disable_shutdown"`
}

func (c *ShutdownConfig) Prepare() []error {
	var errs []error

	if c.Timeout == 0 {
		c.Timeout = 5 * time.Minute
	}

	return errs
}

type StepShutdown struct {
	Config *ShutdownConfig
}

func (s *StepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	if off, _ := vm.IsPoweredOff(); off {
		// Probably power off initiated by last provisioner, though disable_shutdown is not set
		ui.Say("VM is already powered off")
		return multistep.ActionContinue
	}

	if s.Config.DisableShutdown {
		ui.Say("Automatic shutdown disabled. Please shutdown virtual machine.")
	} else if s.Config.Command != "" {
		ui.Say("Executing shutdown command...")
		log.Printf("Shutdown command: %s", s.Config.Command)

		var stdout, stderr bytes.Buffer
		cmd := &packer.RemoteCmd{
			Command: s.Config.Command,
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		err := comm.Start(ctx, cmd)
		if err != nil {
			state.Put("error", fmt.Errorf("Failed to send shutdown command: %s", err))
			return multistep.ActionHalt
		}
	} else {
		ui.Say("Shutting down VM...")

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
