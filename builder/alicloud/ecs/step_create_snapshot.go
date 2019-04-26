package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateAlicloudSnapshot struct {
	snapshot                 *ecs.Snapshot
	WaitSnapshotReadyTimeout int
}

func (s *stepCreateAlicloudSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.Instance)

	describeDisksRequest := ecs.CreateDescribeDisksRequest()
	describeDisksRequest.RegionId = config.AlicloudRegion
	describeDisksRequest.InstanceId = instance.InstanceId
	describeDisksRequest.DiskType = DiskTypeSystem
	disksResponse, err := client.DescribeDisks(describeDisksRequest)
	if err != nil {
		return halt(state, err, "Error describe disks")
	}

	disks := disksResponse.Disks.Disk
	if len(disks) == 0 {
		return halt(state, err, "Unable to find system disk of instance")
	}

	// Create the alicloud snapshot
	ui.Say(fmt.Sprintf("Creating snapshot from system disk: %s", disks[0].DiskId))

	createSnapshotRequest := ecs.CreateCreateSnapshotRequest()
	createSnapshotRequest.DiskId = disks[0].DiskId
	snapshot, err := client.CreateSnapshot(createSnapshotRequest)
	if err != nil {
		return halt(state, err, "Error creating snapshot")
	}

	_, err = client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDescribeSnapshotsRequest()
			request.RegionId = config.AlicloudRegion
			request.SnapshotIds = snapshot.SnapshotId
			return client.DescribeSnapshots(request)
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			snapshotsResponse := response.(*ecs.DescribeSnapshotsResponse)
			snapshots := snapshotsResponse.Snapshots.Snapshot
			for _, snapshot := range snapshots {
				if snapshot.Status == SnapshotStatusAccomplished {
					return WaitForExpectSuccess
				}
			}
			return WaitForExpectToRetry
		},
		RetryTimeout: time.Duration(s.WaitSnapshotReadyTimeout) * time.Second,
	})

	if err != nil {
		return halt(state, err, "Timeout waiting for snapshot to be created")
	}

	describeSnapshotsRequest := ecs.CreateDescribeSnapshotsRequest()
	describeSnapshotsRequest.RegionId = config.AlicloudRegion
	describeSnapshotsRequest.SnapshotIds = snapshot.SnapshotId

	snapshotsResponse, err := client.DescribeSnapshots(describeSnapshotsRequest)
	if err != nil {
		return halt(state, err, "Error querying created snapshot")
	}

	snapshots := snapshotsResponse.Snapshots.Snapshot
	if len(snapshots) == 0 {
		return halt(state, err, "Unable to find created snapshot")
	}

	s.snapshot = &snapshots[0]
	state.Put("alicloudsnapshot", snapshot.SnapshotId)
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

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting the snapshot because of cancellation or error...")

	deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
	deleteSnapshotRequest.SnapshotId = s.snapshot.SnapshotId
	if _, err := client.DeleteSnapshot(deleteSnapshotRequest); err != nil {
		ui.Error(fmt.Sprintf("Error deleting snapshot, it may still be around: %s", err))
		return
	}
}
