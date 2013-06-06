package vmware

import (
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
//   vmx_path string
//
// Produces:
//   <nothing>
type stepShutdown struct{}

func (s *stepShutdown) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	if config.ShutdownCommand != "" {
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", config.ShutdownCommand)
		cmd := &packer.RemoteCmd{Command: config.ShutdownCommand}
		if err := comm.Start(cmd); err != nil {
			ui.Error(fmt.Sprintf("Failed to send shutdown command: %s", err))
			return multistep.ActionHalt
		}

		// Wait for the command to run
		cmd.Wait()

		// Wait for the machine to actually shut down
		log.Printf("Waiting max %s for shutdown to complete", config.ShutdownTimeout)
		shutdownTimer := time.After(config.ShutdownTimeout)
		for {
			running, _ := driver.IsRunning(vmxPath)
			if !running {
				break
			}

			select {
			case <-shutdownTimer:
				ui.Error("Timeout while waiting for machine to shut down.")
				return multistep.ActionHalt
			default:
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		if err := driver.Stop(vmxPath); err != nil {
			ui.Error(fmt.Sprintf("Error stopping VM: %s", err))
			return multistep.ActionHalt
		}
	}

	log.Println("VM shut down.")
	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state map[string]interface{}) {}
