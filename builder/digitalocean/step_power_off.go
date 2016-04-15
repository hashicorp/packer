package digitalocean

import (
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepPowerOff struct{}

func (s *stepPowerOff) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	c := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	dropletId := state.Get("droplet_id").(int)

	droplet, _, err := client.Droplets.Get(dropletId)
	if err != nil {
		err := fmt.Errorf("Error checking droplet state: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if droplet.Status == "off" {
		// Droplet is already off, don't do anything
		return multistep.ActionContinue
	}

	// Pull the plug on the Droplet
	ui.Say("Forcefully shutting down Droplet...")
	_, _, err = client.DropletActions.PowerOff(dropletId)
	if err != nil {
		err := fmt.Errorf("Error powering off droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for poweroff event to complete...")
	err = waitForDropletState("off", dropletId, client, c.StateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the droplet to become unlocked for future steps
	if err := waitForDropletUnlocked(client, dropletId, c.StateTimeout); err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error powering off droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPowerOff) Cleanup(state multistep.StateBag) {
	// no cleanup
}
