package digitalocean

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateVolumes struct {
	volumeIDs []string
}

func (s *stepCreateVolumes) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)

	if len(c.Volumes) == 0 {
		return multistep.ActionContinue
	}

	// Create the volumes based on configuration
	ui.Say("Creating volumes...")

	for _, vc := range c.Volumes {
		volume, _, err := client.Storage.CreateVolume(context.TODO(), &godo.VolumeCreateRequest{
			Name:          vc.VolumeName,
			SizeGigaBytes: vc.Size,
			Region:        c.Region,
			SnapshotID:    vc.BaseSnapshotID,
		})
		if err != nil {
			err := fmt.Errorf("Error creating volume: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		s.volumeIDs = append(s.volumeIDs, volume.ID)
	}

	// Store the droplet id for later
	state.Put("volume_ids", s.volumeIDs)

	return multistep.ActionContinue
}

func (s *stepCreateVolumes) Cleanup(state multistep.StateBag) {
	if len(s.volumeIDs) == 0 {
		return
	}

	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Destroying volumes...")
	for _, volumeID := range s.volumeIDs {
		_, err := client.Storage.DeleteVolume(context.TODO(), volumeID)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error destroying volume %q. Please destroy it manually: %s",
				volumeID, err))
		}
	}
}
