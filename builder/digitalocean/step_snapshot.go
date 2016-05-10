package digitalocean

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/digitalocean/godo"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepSnapshot struct{}

func (s *stepSnapshot) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	dropletId := state.Get("droplet_id").(int)

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.SnapshotName))
	_, _, err := client.DropletActions.Snapshot(dropletId, c.SnapshotName)
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the droplet to become unlocked first. For snapshots
	// this can end up taking quite a long time, so we hardcode this to
	// 20 minutes.
	if err := waitForDropletUnlocked(client, dropletId, 20*time.Minute); err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error shutting down droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// With the pending state over, verify that we're in the active state
	ui.Say("Waiting for snapshot to complete...")
	err = waitForDropletState("active", dropletId, client, c.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for snapshot to complete: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Looking up snapshot ID for snapshot: %s", c.SnapshotName)
	images, _, err := client.Droplets.Snapshots(dropletId, nil)
	if err != nil {
		err := fmt.Errorf("Error looking up snapshot ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var imageId int
	if len(images) == 1 {
		imageId = images[0].ID
	} else {
		err := errors.New("Couldn't find snapshot to get the image ID. Bug?")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Snapshot image ID: %d", imageId)
	state.Put("snapshot_image_id", imageId)
	state.Put("snapshot_name", c.SnapshotName)
	state.Put("region", c.Region)

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}
