package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepShutdown struct{}

func (s *stepShutdown) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*DigitalOceanClient)
	c := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)
	dropletId := state.Get("droplet_id").(uint)

	// Gracefully power off the droplet. We have to retry this a number
	// of times because sometimes it says it completed when it actually
	// did absolutely nothing (*ALAKAZAM!* magic!). We give up after
	// a pretty arbitrary amount of time.
	var err error
	ui.Say("Gracefully shutting down droplet...")
	for attempts := 1; attempts <= 10; attempts++ {
		log.Printf("ShutdownDropetl attempt #%d...", attempts)
		err := client.ShutdownDroplet(dropletId)
		if err != nil {
			err := fmt.Errorf("Error shutting down droplet: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		err = waitForDropletState("off", dropletId, client, c.stateTimeout)
		if err == nil {
			break
		}
	}

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
	// no cleanup
}
