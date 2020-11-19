package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepServerInfo struct{}

func (s *stepServerInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	serverID := state.Get("server_id").(string)

	ui.Say("Waiting for server to become active...")

	instanceResp, err := instanceAPI.WaitForServer(&instance.WaitForServerRequest{
		ServerID: serverID,
	})
	if err != nil {
		err := fmt.Errorf("Error waiting for server to become booted: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instanceResp.State != instance.ServerStateRunning {
		err := fmt.Errorf("Server is in state %s", instanceResp.State.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instanceResp.PublicIP == nil {
		err := fmt.Errorf("Server does not have a public IP")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("server_ip", instanceResp.PublicIP.Address.String())
	state.Put("root_volume_id", instanceResp.Volumes["0"].ID)

	return multistep.ActionContinue
}

func (s *stepServerInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
