package digitalocean

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

type stepPowerOff struct{}

func (s *stepPowerOff) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	dropletId := state["droplet_id"].(uint)

	// Sleep arbitrarily before sending power off request
	// Otherwise we get "pending event" errors, even though there isn't
	// one.
	time.Sleep(3 * time.Second)

	// Poweroff the droplet so it can be snapshot
	err := client.PowerOffDroplet(dropletId)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for droplet to power off...")

	err = waitForDropletState("off", dropletId, client)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPowerOff) Cleanup(state map[string]interface{}) {
	// no cleanup
}
