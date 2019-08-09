package chroot

import (
	"context"
	"fmt"
	"time"

	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepSnapshot creates a snapshot of the created volume.
//
// Produces:
//   snapshot_id string - ID of the created snapshot
type StepSnapshot struct {
	snapshotId string
}

func (s *StepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)
	volumeId := state.Get("volume_id").(string)

	ui.Say("Creating snapshot...")
	description := fmt.Sprintf("Packer: %s", time.Now().String())

	createSnapResp, err := oapiconn.POST_CreateSnapshot(oapi.CreateSnapshotRequest{
		VolumeId:    volumeId,
		Description: description,
	})
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the snapshot ID so we can delete it later
	s.snapshotId = createSnapResp.OK.Snapshot.SnapshotId
	ui.Message(fmt.Sprintf("Snapshot ID: %s", s.snapshotId))

	// Wait for the snapshot to be ready
	err = osccommon.WaitUntilSnapshotDone(oapiconn, s.snapshotId)
	if err != nil {
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshot_id", s.snapshotId)

	snapshots := map[string][]string{
		oapiconn.GetConfig().Region: {s.snapshotId},
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
		oapiconn := state.Get("oapi").(*oapi.Client)
		ui := state.Get("ui").(packer.Ui)
		ui.Say("Removing snapshot since we cancelled or halted...")
		_, err := oapiconn.POST_DeleteSnapshot(oapi.DeleteSnapshotRequest{SnapshotId: s.snapshotId})
		if err != nil {
			ui.Error(fmt.Sprintf("Error: %s", err))
		}
	}
}
