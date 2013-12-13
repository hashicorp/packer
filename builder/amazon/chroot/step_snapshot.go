package chroot

import (
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

// StepSnapshot creates a snapshot of the created volume.
//
// Produces:
//   snapshot_id string - ID of the created snapshot
type StepSnapshot struct {
	snapshotId string
}

func (s *StepSnapshot) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	volumeId := state.Get("volume_id").(string)

	ui.Say("Creating snapshot...")
	createSnapResp, err := ec2conn.CreateSnapshot(volumeId, "")
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the snapshot ID so we can delete it later
	s.snapshotId = createSnapResp.Id
	ui.Message(fmt.Sprintf("Snapshot ID: %s", s.snapshotId))

	// Wait for the snapshot to be ready
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"pending"},
		StepState: state,
		Target:    "completed",
		Refresh: func() (interface{}, string, error) {
			resp, err := ec2conn.Snapshots([]string{s.snapshotId}, ec2.NewFilter())
			if err != nil {
				return nil, "", err
			}

			if len(resp.Snapshots) == 0 {
				return nil, "", errors.New("No snapshots found.")
			}

			s := resp.Snapshots[0]
			return s, s.Status, nil
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
		_, err := ec2conn.DeleteSnapshots([]string{s.snapshotId})
		if err != nil {
			ui.Error(fmt.Sprintf("Error: %s", err))
		}
	}
}
