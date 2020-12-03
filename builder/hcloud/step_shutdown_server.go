package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type stepShutdownServer struct{}

func (s *stepShutdownServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	serverID := state.Get("server_id").(int)

	ui.Say("Shutting down server...")

	action, _, err := client.Server.Shutdown(ctx, &hcloud.Server{ID: serverID})

	if err != nil {
		err := fmt.Errorf("Error stopping server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	_, errCh := client.Action.WatchProgress(ctx, action)
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
