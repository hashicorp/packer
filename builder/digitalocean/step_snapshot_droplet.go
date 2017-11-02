package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepSnapshotDroplet struct{}

func (s *stepSnapshotDroplet) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	dropletId := state.Get("droplet_id").(int)

	ui.Say(fmt.Sprintf("Creating droplet snapshot: %v", c.SnapshotName))
	action, _, err := client.DropletActions.Snapshot(context.TODO(), dropletId, c.SnapshotName)
	if err != nil {
		err := fmt.Errorf("Error creating droplet snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// With the pending state over, verify that we're in the active state
	ui.Say("Waiting for droplet snapshot to complete...")
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
		err := fmt.Errorf("Error waiting for droplet to unlock: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Looking up snapshot ID for droplet snapshot: %s", c.SnapshotName)
	images, _, err := client.Droplets.Snapshots(context.TODO(), dropletId, nil)
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

	snapshotRegions := []string{c.Region}
	if len(c.SnapshotRegions) > 0 {
		seenRegions := map[string]struct{}{c.Region: {}}
		for _, region := range c.SnapshotRegions {
			if _, ok := seenRegions[region]; ok {
				continue
			}
			seenRegions[region] = struct{}{}
			snapshotRegions = append(snapshotRegions, region)

			transferRequest := &godo.ActionRequest{
				"type":   "transfer",
				"region": region,
			}
			imageTransfer, _, err := client.ImageActions.Transfer(context.TODO(), imageId, transferRequest)
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

	log.Printf("Snapshot image ID: %d", imageId)
	state.Put("droplet_snapshot", snapshot{
		id:      strconv.Itoa(imageId),
		name:    c.SnapshotName,
		regions: snapshotRegions,
	})

	return multistep.ActionContinue
}

func (s *stepSnapshotDroplet) Cleanup(state multistep.StateBag) {
	// no cleanup
}
