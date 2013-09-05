package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepPowerOff struct{}

func (s *stepPowerOff) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*DigitalOceanClient)
	c := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)
	dropletId := state.Get("droplet_id").(uint)

	_, status, err := client.DropletStatus(dropletId)
	if err != nil {
		err := fmt.Errorf("Error checking droplet state: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if status == "off" {
		// Droplet is already off, don't do anything
		return multistep.ActionContinue
	}

	// Pull the plug on the Droplet
	ui.Say("Forcefully shutting down Droplet...")
	err = client.PowerOffDroplet(dropletId)
	if err != nil {
		err := fmt.Errorf("Error powering off droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for poweroff event to complete...")
	err = waitForDropletState("off", dropletId, client, c.stateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPowerOff) Cleanup(state multistep.StateBag) {
	// no cleanup
}
