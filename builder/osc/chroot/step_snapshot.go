package chroot

import (
	"context"
	"fmt"
	"time"

	"github.com/antihax/optional"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepSnapshot creates a snapshot of the created volume.
//
// Produces:
//   snapshot_id string - ID of the created snapshot
type StepSnapshot struct {
	snapshotId string
	RawRegion  string
}

func (s *StepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	oscconn := state.Get("osc").(*osc.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
	volumeId := state.Get("volume_id").(string)

	ui.Say("Creating snapshot...")
	description := fmt.Sprintf("Packer: %s", time.Now().String())

	createSnapResp, _, err := oscconn.SnapshotApi.CreateSnapshot(context.Background(), &osc.CreateSnapshotOpts{
		CreateSnapshotRequest: optional.NewInterface(osc.CreateSnapshotRequest{
			VolumeId:    volumeId,
			Description: description,
		}),
	})
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the snapshot ID so we can delete it later
	s.snapshotId = createSnapResp.Snapshot.SnapshotId
	ui.Message(fmt.Sprintf("Snapshot ID: %s", s.snapshotId))

	// Wait for the snapshot to be ready
	err = osccommon.WaitUntilOscSnapshotDone(oscconn, s.snapshotId)
	if err != nil {
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshot_id", s.snapshotId)

	snapshots := map[string][]string{
		s.RawRegion: {s.snapshotId},
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
		oscconn := state.Get("osc").(*osc.APIClient)
		ui := state.Get("ui").(packersdk.Ui)
		ui.Say("Removing snapshot since we cancelled or halted...")
		_, _, err := oscconn.SnapshotApi.DeleteSnapshot(context.Background(), &osc.DeleteSnapshotOpts{
			DeleteSnapshotRequest: optional.NewInterface(osc.DeleteSnapshotRequest{SnapshotId: s.snapshotId}),
		})
		if err != nil {
			ui.Error(fmt.Sprintf("Error: %s", err))
		}
	}
}
