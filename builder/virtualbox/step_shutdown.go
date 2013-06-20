package virtualbox

import (
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
//   config *config
//   driver Driver
//   ui     packer.Ui
//   vmName string
//
// Produces:
//   <nothing>
type stepShutdown struct{}

func (s *stepShutdown) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	if config.ShutdownCommand != "" {
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", config.ShutdownCommand)
		cmd := &packer.RemoteCmd{Command: config.ShutdownCommand}
		if err := comm.Start(cmd); err != nil {
			err := fmt.Errorf("Failed to send shutdown command: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Wait for the command to run
		cmd.Wait()

		// Wait for the machine to actually shut down
		log.Printf("Waiting max %s for shutdown to complete", config.ShutdownTimeout)
		shutdownTimer := time.After(config.ShutdownTimeout)
		for {
			running, _ := driver.IsRunning(vmName)
			if !running {
				break
			}

			select {
			case <-shutdownTimer:
				err := errors.New("Timeout while waiting for machine to shut down.")
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			default:
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		ui.Say("Halting the virtual machine...")
		if err := driver.Stop(vmName); err != nil {
			err := fmt.Errorf("Error stopping VM: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	log.Println("VM shut down.")
	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state map[string]interface{}) {}
