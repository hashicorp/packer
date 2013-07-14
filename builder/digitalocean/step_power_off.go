package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepPowerOff struct{}

func (s *stepPowerOff) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	c := state["config"].(config)
	ui := state["ui"].(packer.Ui)
	dropletId := state["droplet_id"].(uint)

	// Sleep arbitrarily before sending power off request
	// Otherwise we get "pending event" errors, even though there isn't
	// one.
	log.Printf("Sleeping for %v, event_delay", c.RawEventDelay)
	time.Sleep(c.eventDelay)

	// Poweroff the droplet so it can be snapshot
	err := client.PowerOffDroplet(dropletId)

	if err != nil {
		err := fmt.Errorf("Error powering off droplet: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for droplet to power off...")

	err = waitForDropletState("off", dropletId, client, c)
	if err != nil {
		err := fmt.Errorf("Error waiting for droplet to become 'off': %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPowerOff) Cleanup(state map[string]interface{}) {
	// no cleanup
}
