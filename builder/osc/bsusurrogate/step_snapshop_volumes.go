package bsusurrogate

import (
	"context"
	"fmt"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	osccommon "github.com/hashicorp/packer/builder/osc/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepSnapshotVolumes creates snapshots of the created volumes.
//
// Produces:
//   snapshot_ids map[string]string - IDs of the created snapshots
type StepSnapshotVolumes struct {
	LaunchDevices []oapi.BlockDeviceMappingVmCreation
	snapshotIds   map[string]string
}

func (s *StepSnapshotVolumes) snapshotVolume(ctx context.Context, deviceName string, state multistep.StateBag) error {
	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(oapi.Vm)

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

	createSnapResp, err := oapiconn.POST_CreateSnapshot(oapi.CreateSnapshotRequest{
		VolumeId:    volumeId,
		Description: description,
	})
	if err != nil {
		return err
	}

	// Set the snapshot ID so we can delete it later
	s.snapshotIds[deviceName] = createSnapResp.OK.Snapshot.SnapshotId

	// Wait for snapshot to be created
	err = osccommon.WaitUntilSnapshotCompleted(oapiconn, createSnapResp.OK.Snapshot.SnapshotId)
	return err
}

func (s *StepSnapshotVolumes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	s.snapshotIds = map[string]string{}

	var wg sync.WaitGroup
	var errs *multierror.Error
	for _, device := range s.LaunchDevices {
		wg.Add(1)
		go func(device oapi.BlockDeviceMappingVmCreation) {
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
		oapiconn := state.Get("oapi").(*oapi.Client)
		ui := state.Get("ui").(packer.Ui)
		ui.Say("Removing snapshots since we cancelled or halted...")
		for _, snapshotId := range s.snapshotIds {
			_, err := oapiconn.POST_DeleteSnapshot(oapi.DeleteSnapshotRequest{SnapshotId: snapshotId})
			if err != nil {
				ui.Error(fmt.Sprintf("Error: %s", err))
			}
		}
	}
}
