package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepSnapshot struct{}

func (s *stepSnapshot) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	c := state["config"].(config)
	dropletId := state["droplet_id"].(uint)

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.SnapshotName))

	err := client.CreateSnapshot(dropletId, c.SnapshotName)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for snapshot to complete...")

	err = waitForDropletState("active", dropletId, client)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state map[string]interface{}) {
	// no cleanup
}
