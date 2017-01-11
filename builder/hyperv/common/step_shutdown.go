package common

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

// This step shuts down the machine. It first attempts to do so gracefully,
// but ultimately forcefully shuts it down if that fails.
//
// Uses:
//   communicator packer.Communicator
//   dir OutputDir
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepShutdown struct {
	Command string
	Timeout time.Duration
}

func (s *StepShutdown) Run(state multistep.StateBag) multistep.StepAction {

	comm := state.Get("communicator").(packer.Communicator)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

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

		// Wait for the command to run so we can print std{err,out}
		// We don't care if the command errored, since we'll notice
		// if the vm didn't shut down.
		cmd.Wait()

		log.Printf("Shutdown stdout: %s", stdout.String())
		log.Printf("Shutdown stderr: %s", stderr.String())

		// Wait for the machine to actually shut down
		log.Printf("Waiting max %s for shutdown to complete", s.Timeout)
		shutdownTimer := time.After(s.Timeout)
		for {
			running, _ := driver.IsRunning(vmName)
			if !running {
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
	} else {
		ui.Say("Forcibly halting virtual machine...")
		if err := driver.Stop(vmName); err != nil {
			err := fmt.Errorf("Error stopping VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	log.Println("VM shut down.")
	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {}
