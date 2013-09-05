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
	ui := state.Get("ui").(packer.Ui)
	dropletId := state.Get("droplet_id").(uint)

	// Poweroff the droplet so it can be snapshot
	err := client.PowerOffDroplet(dropletId)
	if err != nil {
		err := fmt.Errorf("Error powering off droplet: %s", err)
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
