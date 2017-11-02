package digitalocean

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepSnapshotVolumes struct{}

func (s *stepSnapshotVolumes) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	volumeIDs := state.Get("volume_ids").([]string)

	volumeSnapshots := []snapshot{}
	for i := range volumeIDs {
		snap, _, err := client.Storage.CreateSnapshot(context.TODO(), &godo.SnapshotCreateRequest{
			Name:     c.Volumes[i].SnapshotName,
			VolumeID: volumeIDs[i],
		})
		if err != nil {
			err := fmt.Errorf("Error creating snapshot for volume %d: %s", i, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		volumeSnapshots = append(volumeSnapshots, snapshot{
			id:      snap.ID,
			name:    snap.Name,
			regions: []string{c.Region},
		})
		log.Printf("Volume %d snapshot ID: %s", i, snap.ID)
	}

	state.Put("volume_snapshots", volumeSnapshots)

	return multistep.ActionContinue
}

func (s *stepSnapshotVolumes) Cleanup(state multistep.StateBag) {
	// no cleanup
}
