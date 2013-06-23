package digitalocean

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
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
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for snapshot to complete...")
	err = waitForDropletState("active", dropletId, client, c)
	if err != nil {
		err := fmt.Errorf("Error waiting for snapshot to complete: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Looking up snapshot ID for snapshot: %s", c.SnapshotName)
	images, err := client.Images()
	if err != nil {
		err := fmt.Errorf("Error looking up snapshot ID: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var imageId uint
	for _, image := range images {
		if image.Name == c.SnapshotName {
			imageId = image.Id
			break
		}
	}

	if imageId == 0 {
		err := errors.New("Couldn't find snapshot to get the image ID. Bug?")
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Snapshot image ID: %d", imageId)

	state["snapshot_image_id"] = imageId
	state["snapshot_name"] = c.SnapshotName

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state map[string]interface{}) {
	// no cleanup
}
