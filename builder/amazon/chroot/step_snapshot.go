package chroot

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepSnapshot creates a snapshot of the created volume.
//
// Produces:
//   snapshot_id string - ID of the created snapshot
type StepSnapshot struct {
	snapshotId string
}

func (s *StepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	volumeId := state.Get("volume_id").(string)

	ui.Say("Creating snapshot...")
	description := fmt.Sprintf("Packer: %s", time.Now().String())

	createSnapResp, err := ec2conn.CreateSnapshot(&ec2.CreateSnapshotInput{
		VolumeId:    &volumeId,
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
	err = awscommon.WaitUntilSnapshotDone(ctx, ec2conn, s.snapshotId)
	if err != nil {
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshot_id", s.snapshotId)

	snapshots := map[string][]string{
		*ec2conn.Config.Region: {s.snapshotId},
	}
	state.Put("snapshots", snapshots)

	return multistep.ActionContinue
}

func (s *StepSnapshot) Cleanup(state multistep.StateBag) {
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
