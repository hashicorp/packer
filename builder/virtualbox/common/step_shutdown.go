package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step shuts down the machine. It first attempts to do so gracefully,
// but ultimately forcefully shuts it down if that fails.
//
// Uses:
//   communicator packer.Communicator
//   driver Driver
//   ui     packer.Ui
//   vmName string
//
// Produces:
//   <nothing>
type StepShutdown struct {
	Command         string
	Timeout         time.Duration
	Delay           time.Duration
	DisableShutdown bool
}

func (s *StepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if !s.DisableShutdown {
		if s.Command != "" {
			ui.Say("Gracefully halting virtual machine...")
			log.Printf("Executing shutdown command: %s", s.Command)
			cmd := &packer.RemoteCmd{Command: s.Command}
			if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
				err := fmt.Errorf("Failed to send shutdown command: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

		} else {
			ui.Say("Halting the virtual machine...")
			if err := driver.Stop(vmName); err != nil {
				err := fmt.Errorf("Error stopping VM: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	} else {
		ui.Say("Automatic shutdown disabled. Please shutdown virtual machine.")
	}

	// Wait for the machine to actually shut down
	log.Printf("Waiting max %s for shutdown to complete", s.Timeout)
	shutdownTimer := time.After(s.Timeout)
	for {
		running, _ := driver.IsRunning(vmName)
		if !running {

			if s.Delay.Nanoseconds() > 0 {
				log.Printf("Delay for %s after shutdown to allow locks to clear...", s.Delay)
				time.Sleep(s.Delay)
			}

			break
		}

		select {
		case <-shutdownTimer:
			err := errors.New("Timeout while waiting for machine to shutdown.")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Println("VM shut down.")
	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {}
