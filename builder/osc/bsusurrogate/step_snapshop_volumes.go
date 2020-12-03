package bsusurrogate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/antihax/optional"
	multierror "github.com/hashicorp/go-multierror"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepSnapshotVolumes creates snapshots of the created volumes.
//
// Produces:
//   snapshot_ids map[string]string - IDs of the created snapshots
type StepSnapshotVolumes struct {
	LaunchDevices []osc.BlockDeviceMappingVmCreation
	snapshotIds   map[string]string
}

func (s *StepSnapshotVolumes) snapshotVolume(ctx context.Context, deviceName string, state multistep.StateBag) error {
	oscconn := state.Get("osc").(*osc.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(osc.Vm)

	var volumeId string
	for _, volume := range vm.BlockDeviceMappings {
		if volume.DeviceName == deviceName {
			volumeId = volume.Bsu.VolumeId
		}
	}
	if volumeId == "" {
		return fmt.Errorf("Volume ID for device %s not found", deviceName)
	}

	ui.Say(fmt.Sprintf("Creating snapshot of EBS Volume %s...", volumeId))
	description := fmt.Sprintf("Packer: %s", time.Now().String())

	createSnapResp, _, err := oscconn.SnapshotApi.CreateSnapshot(context.Background(), &osc.CreateSnapshotOpts{
		CreateSnapshotRequest: optional.NewInterface(osc.CreateSnapshotRequest{
			VolumeId:    volumeId,
			Description: description,
		}),
	})
	if err != nil {
		return err
	}

	// Set the snapshot ID so we can delete it later
	s.snapshotIds[deviceName] = createSnapResp.Snapshot.SnapshotId

	// Wait for snapshot to be created
	err = osccommon.WaitUntilOscSnapshotCompleted(oscconn, createSnapResp.Snapshot.SnapshotId)
	return err
}

func (s *StepSnapshotVolumes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	s.snapshotIds = map[string]string{}

	var wg sync.WaitGroup
	var errs *multierror.Error
	for _, device := range s.LaunchDevices {
		wg.Add(1)
		go func(device osc.BlockDeviceMappingVmCreation) {
			defer wg.Done()
			if err := s.snapshotVolume(ctx, device.DeviceName, state); err != nil {
				errs = multierror.Append(errs, err)
			}
		}(device)
	}

	wg.Wait()

	if errs != nil {
		state.Put("error", errs)
		ui.Error(errs.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshot_ids", s.snapshotIds)
	return multistep.ActionContinue
}

func (s *StepSnapshotVolumes) Cleanup(state multistep.StateBag) {
	if len(s.snapshotIds) == 0 {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		oscconn := state.Get("osc").(*osc.APIClient)
		ui := state.Get("ui").(packersdk.Ui)
		ui.Say("Removing snapshots since we cancelled or halted...")
		for _, snapshotID := range s.snapshotIds {
			_, _, err := oscconn.SnapshotApi.DeleteSnapshot(context.Background(), &osc.DeleteSnapshotOpts{
				DeleteSnapshotRequest: optional.NewInterface(osc.DeleteSnapshotRequest{SnapshotId: snapshotID}),
			})
			if err != nil {
				ui.Error(fmt.Sprintf("Error: %s", err))
			}
		}
	}
}
