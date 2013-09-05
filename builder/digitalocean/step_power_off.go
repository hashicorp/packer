package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepPowerOff struct{}

func (s *stepPowerOff) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*DigitalOceanClient)
	ui := state.Get("ui").(packer.Ui)
	dropletId := state.Get("droplet_id").(uint)

	// Gracefully power off the droplet. We have to retry this a number
	// of times because sometimes it says it completed when it actually
	// did absolutely nothing (*ALAKAZAM!* magic!). We give up after
	// a pretty arbitrary amount of time.
	var err error
	ui.Say("Gracefully shutting down droplet...")
	for attempts := 1; attempts <= 10; attempts++ {
		log.Printf("PowerOffDroplet attempt #%d...", attempts)
		err := client.PowerOffDroplet(dropletId)
		if err != nil {
			err := fmt.Errorf("Error powering off droplet: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		err = waitForDropletState("off", dropletId, client, 20*time.Second)
		if err == nil {
			// We reached the state!
			break
		}
	}

	if err != nil {
		err := fmt.Errorf("Error waiting for droplet to become 'off': %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for poweroff event to complete...")
	return multistep.ActionContinue
}

func (s *stepPowerOff) Cleanup(state multistep.StateBag) {
	// no cleanup
}
