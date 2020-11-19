package qemu

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step shuts down the machine. It first attempts to do so gracefully,
// but ultimately forcefully shuts it down if that fails.
//
// Uses:
//   communicator packer.Communicator
//   config *config
//   driver Driver
//   ui     packersdk.Ui
//
// Produces:
//   <nothing>
type stepShutdown struct {
	ShutdownCommand string
	ShutdownTimeout time.Duration
	Comm            *communicator.Config
}

func (s *stepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if s.Comm.Type == "none" {
		cancelCh := make(chan struct{}, 1)
		go func() {
			defer close(cancelCh)
			<-time.After(s.ShutdownTimeout)
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

	if s.ShutdownCommand != "" {
		comm := state.Get("communicator").(packer.Communicator)
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", s.ShutdownCommand)
		cmd := &packer.RemoteCmd{Command: s.ShutdownCommand}
		if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
			err := fmt.Errorf("Failed to send shutdown command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Start the goroutine that will time out our graceful attempt
		cancelCh := make(chan struct{}, 1)
		go func() {
			defer close(cancelCh)
			<-time.After(s.ShutdownTimeout)
		}()

		log.Printf("Waiting max %s for shutdown to complete", s.ShutdownTimeout)
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
