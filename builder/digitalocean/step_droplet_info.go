package digitalocean

import (
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDropletInfo struct{}

func (s *stepDropletInfo) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	dropletId := state.Get("droplet_id").(int)

	ui.Say("Waiting for droplet to become active...")

	err := waitForDropletState("active", dropletId, client, c.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for droplet to become active: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the IP on the state for later
	droplet, _, err := client.Droplets.Get(dropletId)
	if err != nil {
		err := fmt.Errorf("Error retrieving droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Verify we have an IPv4 address
	invalid := droplet.Networks == nil ||
		len(droplet.Networks.V4) == 0
	if invalid {
		err := fmt.Errorf("IPv4 address not found for droplet!")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("droplet_ip", droplet.Networks.V4[0].IPAddress)
	return multistep.ActionContinue
}

func (s *stepDropletInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}
