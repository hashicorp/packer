package openstack

import (
	"context"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepWaitForRackConnect struct {
	Wait bool
}

func (s *StepWaitForRackConnect) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.Wait {
		return multistep.ActionContinue
	}

	config := state.Get("config").(*Config)
	server := state.Get("server").(*servers.Server)
	ui := state.Get("ui").(packersdk.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf(
		"Waiting for server (%s) to become RackConnect ready...", server.ID))
	for {
		server, err = servers.Get(computeClient, server.ID).Extract()
		if err != nil {
			return multistep.ActionHalt
		}

		if server.Metadata["rackconnect_automation_status"] == "DEPLOYED" {
			state.Put("server", server)
			break
		}

		time.Sleep(2 * time.Second)
	}

	return multistep.ActionContinue
}

func (s *StepWaitForRackConnect) Cleanup(state multistep.StateBag) {
}
