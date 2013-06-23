package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDropletInfo struct{}

func (s *stepDropletInfo) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	c := state["config"].(config)
	dropletId := state["droplet_id"].(uint)

	ui.Say("Waiting for droplet to become active...")

	err := waitForDropletState("active", dropletId, client, c)
	if err != nil {
		err := fmt.Errorf("Error waiting for droplet to become active: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the IP on the state for later
	ip, _, err := client.DropletStatus(dropletId)
	if err != nil {
		err := fmt.Errorf("Error retrieving droplet ID: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["droplet_ip"] = ip

	return multistep.ActionContinue
}

func (s *stepDropletInfo) Cleanup(state map[string]interface{}) {
	// no cleanup
}
