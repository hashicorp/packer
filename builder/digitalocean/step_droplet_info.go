package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDropletInfo struct{}

func (s *stepDropletInfo) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*DigitalOceanClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	dropletId := state.Get("droplet_id").(uint)

	ui.Say("Waiting for droplet to become active...")

	err := waitForDropletState("active", dropletId, client, c.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for droplet to become active: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the IP on the state for later
	ip, _, err := client.DropletStatus(dropletId)
	if err != nil {
		err := fmt.Errorf("Error retrieving droplet ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("droplet_ip", ip)

	return multistep.ActionContinue
}

func (s *stepDropletInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
