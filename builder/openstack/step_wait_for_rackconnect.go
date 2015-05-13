package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"

	"github.com/mitchellh/gophercloud-fork-40444fb"
)

type StepWaitForRackConnect struct {
	Wait bool
}

func (s *StepWaitForRackConnect) Run(state multistep.StateBag) multistep.StepAction {
	if !s.Wait {
		return multistep.ActionContinue
	}

	csp := state.Get("csp").(gophercloud.CloudServersProvider)
	server := state.Get("server").(*gophercloud.Server)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Waiting for server (%s) to become RackConnect ready...", server.Id))

	for {
		server, err := csp.ServerById(server.Id)
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
