package qemu

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
//   config *config
//   driver Driver
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepShutdown struct{}

func (s *stepShutdown) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if state.Get("communicator") == nil {
		cancelCh := make(chan struct{}, 1)
		go func() {
			defer close(cancelCh)
			<-time.After(config.shutdownTimeout)
		}()
		ui.Say("Waiting for shutdown...")
		if ok := driver.WaitForShutdown(cancelCh); ok {
			log.Println("VM shut down.")
			return multistep.ActionContinue
		} else {
			err := fmt.Errorf("Failed to shutdown")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	comm := state.Get("communicator").(packer.Communicator)
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
