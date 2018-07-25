package scaleway

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type stepSnapshot struct{}

func (s *stepSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	volumeID := state.Get("root_volume_id").(string)

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.SnapshotName))
	snapshot, err := client.PostSnapshot(volumeID, c.SnapshotName)
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Snapshot ID: %s", snapshot)
	state.Put("snapshot_id", snapshot)
	state.Put("snapshot_name", c.SnapshotName)
	state.Put("region", c.Region)

	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}
