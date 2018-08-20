package scaleway

import (
	"context"
	"fmt"

	"bytes"
	"errors"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
	"log"
	"time"
)

type stepShutdown struct {
	Command string
	Timeout time.Duration
}

func (s *stepShutdown) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	comm := state.Get("communicator").(packer.Communicator)
	ui := state.Get("ui").(packer.Ui)
	serverID := state.Get("server_id").(string)

	ui.Say("Shutting down server...")

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

		// Wait for the machine to actually shut down
		log.Printf("Waiting max %s for shutdown to complete", s.Timeout)
		shutdownTimer := time.After(s.Timeout)
		for {
			select {
			case <-shutdownTimer:
				log.Printf("Shutdown stdout: %s", stdout.String())
				log.Printf("Shutdown stderr: %s", stderr.String())
				err := errors.New("Timeout while waiting for machine to shut down.")
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			default:
				time.Sleep(500 * time.Millisecond)
			}
		}
	} else {
		ui.Say("Forcibly halting virtual machine...")
		err := client.PostServerAction(serverID, "poweroff")

		if err != nil {
			err := fmt.Errorf("Error stopping server: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		_, err = api.WaitForServerState(client, serverID, "stopped")

		if err != nil {
			err := fmt.Errorf("Error shutting down server: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
	// no cleanup
}
