package qemu

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
//
// Produces:
//   <nothing>
type stepShutdown struct{}

func (s *stepShutdown) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if config.ShutdownCommand != "" {
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", config.ShutdownCommand)
		cmd := &packer.RemoteCmd{Command: config.ShutdownCommand}
		if err := cmd.StartWithUi(comm, ui); err != nil {
			err := fmt.Errorf("Failed to send shutdown command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Start the goroutine that will time out our graceful attempt
		cancelCh := make(chan struct{}, 1)
		go func() {
			defer close(cancelCh)
			<-time.After(config.shutdownTimeout)
		}()

		log.Printf("Waiting max %s for shutdown to complete", config.shutdownTimeout)
		if ok := driver.WaitForShutdown(cancelCh); !ok {
			err := errors.New("Timeout while waiting for machine to shut down.")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		ui.Say("Halting the virtual machine...")
		if err := driver.Stop(); err != nil {
			err := fmt.Errorf("Error stopping VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	log.Println("VM shut down.")
	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {}
