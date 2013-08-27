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

	log.Println("Waiting for poweroff event to complete...")

	// This arbitrary sleep is because we can't wait for the state
	// of the droplet to be 'off', as stepShutdown should already
	// have accomplished that, and the state indicator is the same.
	// We just have to assume that this event will process quickly.
	log.Printf("Sleeping for %v, event_delay", c.RawEventDelay)
	time.Sleep(c.eventDelay)

	return multistep.ActionContinue
}

func (s *stepPowerOff) Cleanup(state map[string]interface{}) {
	// no cleanup
}
