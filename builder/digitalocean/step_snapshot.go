package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepSnapshot struct{}

func (s *stepSnapshot) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	dropletId := state.Get("droplet_id").(int)
	var snapshotRegions []string

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.SnapshotName))
	action, _, err := client.DropletActions.Snapshot(context.TODO(), dropletId, c.SnapshotName)
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// With the pending state over, verify that we're in the active state
	ui.Say("Waiting for snapshot to complete...")
	if err := waitForActionState(godo.ActionCompleted, dropletId, action.ID,
		client, 20*time.Minute); err != nil {
		// If we get an error the first time, actually report it
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
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

	log.Printf("Looking up snapshot ID for snapshot: %s", c.SnapshotName)
	images, _, err := client.Droplets.Snapshots(context.TODO(), dropletId, nil)
	if err != nil {
		err := fmt.Errorf("Error looking up snapshot ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(c.SnapshotRegions) > 0 {
		regionSet := make(map[string]struct{})
		regions := make([]string, 0, len(c.SnapshotRegions))
		regionSet[c.Region] = struct{}{}
		for _, region := range c.SnapshotRegions {
			// If we already saw the region, then don't look again
			if _, ok := regionSet[region]; ok {
				continue
			}

			// Mark that we saw the region
			regionSet[region] = struct{}{}

			regions = append(regions, region)
		}
		snapshotRegions = regions

		for transfer := range snapshotRegions {
			transferRequest := &godo.ActionRequest{
				"type":   "transfer",
				"region": snapshotRegions[transfer],
			}
			imageTransfer, _, err := client.ImageActions.Transfer(context.TODO(), images[0].ID, transferRequest)
			if err != nil {
				err := fmt.Errorf("Error transferring snapshot: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			ui.Say(fmt.Sprintf("transferring Snapshot ID: %d", imageTransfer.ID))
			if err := waitForImageState(godo.ActionCompleted, imageTransfer.ID, action.ID,
				client, 20*time.Minute); err != nil {
				// If we get an error the first time, actually report it
				err := fmt.Errorf("Error waiting for snapshot transfer: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
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
	state.Put("regions", snapshotRegions)

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}
