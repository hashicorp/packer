package openstack_id3

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type StepWaitForRackConnect struct {
	Wait bool
}

func (s *StepWaitForRackConnect) Run(state multistep.StateBag) multistep.StepAction {
	if !s.Wait {
		return multistep.ActionContinue
	}

	computeClient := state.Get("compute_client").(*gophercloud.ServiceClient)
	server := state.Get("server").(*servers.Server)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Waiting for server (%s) to become RackConnect ready...", server.ID))

	for {
		server, err := servers.Get(computeClient, server.ID).Extract()
		if err != nil {
			return multistep.ActionHalt
		}

		if server.Metadata["rackconnect_automation_status"] == "DEPLOYED" {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return multistep.ActionContinue
}

func (s *StepWaitForRackConnect) Cleanup(state multistep.StateBag) {
}
