package digitalocean

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDestroyDroplet struct{}

func (s *stepDestroyDroplet) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	dropletId := state["droplet_id"].(uint)

	ui.Say("Destroying droplet...")

	err := client.DestroyDroplet(dropletId)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDestroyDroplet) Cleanup(state map[string]interface{}) {
	// no cleanup
}
