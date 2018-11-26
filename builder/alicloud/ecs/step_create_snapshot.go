package ecs

import (
	"context"
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateAlicloudSnapshot struct {
	snapshot                 *ecs.SnapshotType
	WaitSnapshotReadyTimeout int
}

func (s *stepCreateAlicloudSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	disks, _, err := client.DescribeDisks(&ecs.DescribeDisksArgs{
		RegionId:   common.Region(config.AlicloudRegion),
		InstanceId: instance.InstanceId,
		DiskType:   ecs.DiskTypeAllSystem,
	})

	if err != nil {
		return halt(state, err, "Error describe disks")
	}
	if len(disks) == 0 {
		return halt(state, err, "Unable to find system disk of instance")
	}

	// Create the alicloud snapshot
	ui.Say(fmt.Sprintf("Creating snapshot from system disk: %s", disks[0].DiskId))

	snapshotId, err := client.CreateSnapshot(&ecs.CreateSnapshotArgs{
		DiskId: disks[0].DiskId,
	})

	if err != nil {
		return halt(state, err, "Error creating snapshot")
	}

	err = client.WaitForSnapShotReady(common.Region(config.AlicloudRegion), snapshotId, s.WaitSnapshotReadyTimeout)
	if err != nil {
		return halt(state, err, "Timeout waiting for snapshot to be created")
	}

	snapshots, _, err := client.DescribeSnapshots(&ecs.DescribeSnapshotsArgs{
		RegionId:    common.Region(config.AlicloudRegion),
		SnapshotIds: []string{snapshotId},
	})

	if err != nil {
		return halt(state, err, "Error querying created snapshot")
	}
	if len(snapshots) == 0 {
		return halt(state, err, "Unable to find created snapshot")
	}
	s.snapshot = &snapshots[0]
	state.Put("alicloudsnapshot", snapshotId)

	return multistep.ActionContinue
}

func (s *stepCreateAlicloudSnapshot) Cleanup(state multistep.StateBag) {
	if s.snapshot == nil {
		return
	}
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting the snapshot because of cancellation or error...")
	if err := client.DeleteSnapshot(s.snapshot.SnapshotId); err != nil {
		ui.Error(fmt.Sprintf("Error deleting snapshot, it may still be around: %s", err))
		return
	}
}
