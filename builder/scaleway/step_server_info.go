package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type stepServerInfo struct{}

func (s *stepServerInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	serverID := state.Get("server_id").(string)

	ui.Say("Waiting for server to become active...")

	_, err := api.WaitForServerState(client, serverID, "running")
	if err != nil {
		err := fmt.Errorf("Error waiting for server to become booted: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	server, err := client.GetServer(serverID)
	if err != nil {
		err := fmt.Errorf("Error retrieving server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("server_ip", server.PublicAddress.IP)
	state.Put("root_volume_id", server.Volumes["0"].Identifier)

	return multistep.ActionContinue
}

func (s *stepServerInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
