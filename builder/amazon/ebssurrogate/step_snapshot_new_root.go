package ebssurrogate

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepSnapshotNewRootVolume creates a snapshot of the created volume.
//
// Produces:
//   snapshot_id string - ID of the created snapshot
type StepSnapshotNewRootVolume struct {
	NewRootMountPoint string
	snapshotId        string
}

func (s *StepSnapshotNewRootVolume) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ec2.Instance)

	var newRootVolume string
	for _, volume := range instance.BlockDeviceMappings {
		if *volume.DeviceName == s.NewRootMountPoint {
			newRootVolume = *volume.Ebs.VolumeId
		}
	}

	ui.Say(fmt.Sprintf("Creating snapshot of EBS Volume %s...", newRootVolume))
	description := fmt.Sprintf("Packer: %s", time.Now().String())

	createSnapResp, err := ec2conn.CreateSnapshot(&ec2.CreateSnapshotInput{
		VolumeId:    &newRootVolume,
		Description: &description,
	})
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the snapshot ID so we can delete it later
	s.snapshotId = *createSnapResp.SnapshotId
	ui.Message(fmt.Sprintf("Snapshot ID: %s", s.snapshotId))

	// Wait for the snapshot to be ready
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"pending"},
		StepState: state,
		Target:    "completed",
		Refresh: func() (interface{}, string, error) {
			resp, err := ec2conn.DescribeSnapshots(&ec2.DescribeSnapshotsInput{SnapshotIds: []*string{&s.snapshotId}})
			if err != nil {
				return nil, "", err
			}

			if len(resp.Snapshots) == 0 {
				return nil, "", errors.New("No snapshots found.")
			}

			s := resp.Snapshots[0]
			return s, *s.State, nil
		},
	}

	_, err = awscommon.WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshot_id", s.snapshotId)
	return multistep.ActionContinue
}

func (s *StepSnapshotNewRootVolume) Cleanup(state multistep.StateBag) {
	if s.snapshotId == "" {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		ec2conn := state.Get("ec2").(*ec2.EC2)
		ui := state.Get("ui").(packer.Ui)
		ui.Say("Removing snapshot since we cancelled or halted...")
		_, err := ec2conn.DeleteSnapshot(&ec2.DeleteSnapshotInput{SnapshotId: &s.snapshotId})
		if err != nil {
			ui.Error(fmt.Sprintf("Error: %s", err))
		}
	}
}
