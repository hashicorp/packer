package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepShutdown struct{}

func (s *stepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packer.Ui)
	serverID := state.Get("server_id").(string)

	ui.Say("Shutting down server...")

	_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
		Action:   instance.ServerActionPoweroff,
		ServerID: serverID,
	})
	if err != nil {
		err := fmt.Errorf("Error stopping server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	instanceResp, err := instanceAPI.WaitForServer(&instance.WaitForServerRequest{
		ServerID: serverID,
	})
	if err != nil {
		err := fmt.Errorf("Error shutting down server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instanceResp.State != instance.ServerStateStopped {
		err := fmt.Errorf("Server is in state %s instead of stopped", instanceResp.State.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
	// no cleanup
}
