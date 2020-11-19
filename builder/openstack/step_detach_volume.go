package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumeactions"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepDetachVolume struct {
	UseBlockStorageVolume bool
}

func (s *StepDetachVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Proceed only if block storage volume is used.
	if !s.UseBlockStorageVolume {
		return multistep.ActionContinue
	}

	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	blockStorageClient, err := config.blockStorageV3Client()
	if err != nil {
		err = fmt.Errorf("Error initializing block storage client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	volume := state.Get("volume_id").(string)
	ui.Say(fmt.Sprintf("Detaching volume %s (volume id: %s)", config.VolumeName, volume))
	if err := volumeactions.Detach(blockStorageClient, volume, volumeactions.DetachOpts{}).ExtractErr(); err != nil {
		err = fmt.Errorf("Error detaching block storage volume: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Wait for volume to become available.
	ui.Say(fmt.Sprintf("Waiting for volume %s (volume id: %s) to become available...", config.VolumeName, volume))
	if err := WaitForVolume(blockStorageClient, volume); err != nil {
		err := fmt.Errorf("Error waiting for volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepDetachVolume) Cleanup(multistep.StateBag) {
	// No cleanup.
}
