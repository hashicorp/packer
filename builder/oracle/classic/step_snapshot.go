package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepSnapshot struct {
	cleanupSnap bool
}

func (s *stepSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	s.cleanupSnap = false
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Snapshot...")
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	instanceID := state.Get("instance_id").(string)

	// get instances client
	snapshotClient := client.Snapshots()

	// Instances Input
	snapshotInput := &compute.CreateSnapshotInput{
		Instance:     fmt.Sprintf("%s/%s", config.ImageName, instanceID),
		MachineImage: config.ImageName,
		Timeout:      config.SnapshotTimeout,
	}

	snap, err := snapshotClient.CreateSnapshot(snapshotInput)
	if err != nil {
		err = fmt.Errorf("Problem creating snapshot: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.cleanupSnap = true
	state.Put("snapshot", snap)
	ui.Message(fmt.Sprintf("Created snapshot: %s.", snap.Name))
	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// Delete the snapshot
	ui := state.Get("ui").(packer.Ui)
	if !s.cleanupSnap {
		return
	}
	ui.Say("Deleting Snapshot...")
	client := state.Get("client").(*compute.ComputeClient)
	snap := state.Get("snapshot").(*compute.Snapshot)
	snapClient := client.Snapshots()
	snapInput := compute.DeleteSnapshotInput{
		Snapshot:     snap.Name,
		MachineImage: snap.MachineImage,
	}

	err := snapClient.DeleteSnapshotResourceOnly(&snapInput)
	if err != nil {
		err = fmt.Errorf("Problem deleting snapshot: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
	}
	return
}
