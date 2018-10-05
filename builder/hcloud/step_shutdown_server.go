package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type stepShutdownServer struct{}

func (s *stepShutdownServer) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packer.Ui)
	serverID := state.Get("server_id").(int)

	ui.Say("Shutting down server...")

	action, _, err := client.Server.Shutdown(context.TODO(), &hcloud.Server{ID: serverID})

	if err != nil {
		err := fmt.Errorf("Error stopping server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	_, errCh := client.Action.WatchProgress(context.TODO(), action)
	for {
		select {
		case err1 := <-errCh:
			if err1 == nil {
				return multistep.ActionContinue
			} else {
				err := fmt.Errorf("Error stopping server: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

		}
	}
}

func (s *stepShutdownServer) Cleanup(state multistep.StateBag) {
	// no cleanup
}
