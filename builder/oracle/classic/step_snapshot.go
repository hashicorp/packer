package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepSnapshot struct {
	cleanupSnap bool
}

func (s *stepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating Snapshot...")
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.Client)
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
	state.Put("snapshot", snap)
	state.Put("machine_image", snap.MachineImage)
	ui.Message(fmt.Sprintf("Created snapshot: %s.", snap.Name))
	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// Delete the snapshot
	var snap *compute.Snapshot
	if snapshot, ok := state.GetOk("snapshot"); ok {
		snap = snapshot.(*compute.Snapshot)
	} else {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Deleting Snapshot...")
	client := state.Get("client").(*compute.Client)
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
