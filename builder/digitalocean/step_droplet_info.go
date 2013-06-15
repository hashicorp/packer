package digitalocean

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDropletInfo struct{}

func (s *stepDropletInfo) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	dropletId := state["droplet_id"].(uint)

	ui.Say("Waiting for droplet to become active...")

	err := waitForDropletState("active", dropletId, client)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the IP on the state for later
	ip, _, err := client.DropletStatus(dropletId)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["droplet_ip"] = ip

	return multistep.ActionContinue
}

func (s *stepDropletInfo) Cleanup(state map[string]interface{}) {
	// no cleanup
}
