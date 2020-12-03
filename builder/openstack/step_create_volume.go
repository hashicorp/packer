package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCreateVolume struct {
	UseBlockStorageVolume  bool
	VolumeName             string
	VolumeType             string
	VolumeAvailabilityZone string
	volumeID               string
	doCleanup              bool
}

func (s *StepCreateVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Proceed only if block storage volume is required.
	if !s.UseBlockStorageVolume {
		return multistep.ActionContinue
	}

	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	sourceImage := state.Get("source_image").(string)

	// We will need Block Storage and Image services clients.
	blockStorageClient, err := config.blockStorageV3Client()
	if err != nil {
		err = fmt.Errorf("Error initializing block storage client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	volumeSize := config.VolumeSize

	// Get needed volume size from the source image.
	if volumeSize == 0 {
		imageClient, err := config.imageV2Client()
		if err != nil {
			err = fmt.Errorf("Error initializing image client: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		}

		volumeSize, err = GetVolumeSize(imageClient, sourceImage)
		if err != nil {
			err := fmt.Errorf("Error creating volume: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Creating volume...")
	volumeOpts := volumes.CreateOpts{
		Size:             volumeSize,
		VolumeType:       s.VolumeType,
		AvailabilityZone: s.VolumeAvailabilityZone,
		Name:             s.VolumeName,
		ImageID:          sourceImage,
		Metadata:         config.ImageMetadata,
	}
	volume, err := volumes.Create(blockStorageClient, volumeOpts).Extract()
	if err != nil {
		err := fmt.Errorf("Error creating volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for volume to become available.
	ui.Say(fmt.Sprintf("Waiting for volume %s (volume id: %s) to become available...", config.VolumeName, volume.ID))
	if err := WaitForVolume(blockStorageClient, volume.ID); err != nil {
		err := fmt.Errorf("Error waiting for volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Volume was created, so remember to clean it up.
	s.doCleanup = true

	// Set the Volume ID in the state.
	ui.Message(fmt.Sprintf("Volume ID: %s", volume.ID))
	state.Put("volume_id", volume.ID)
	s.volumeID = volume.ID

	return multistep.ActionContinue
}

func (s *StepCreateVolume) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}

	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	blockStorageClient, err := config.blockStorageV3Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up volume. Please delete the volume manually: %s", s.volumeID))
		return
	}

	// Wait for volume to become available.
	status, err := GetVolumeStatus(blockStorageClient, s.volumeID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error getting the volume information. Please delete the volume manually: %s", s.volumeID))
		return
	}

	if status != "available" {
		ui.Say(fmt.Sprintf(
			"Waiting for volume %s (volume id: %s) to become available...", s.VolumeName, s.volumeID))
		if err := WaitForVolume(blockStorageClient, s.volumeID); err != nil {
			ui.Error(fmt.Sprintf(
				"Error getting the volume information. Please delete the volume manually: %s", s.volumeID))
			return
		}
	}
	ui.Say(fmt.Sprintf("Deleting volume: %s ...", s.volumeID))
	err = volumes.Delete(blockStorageClient, s.volumeID, volumes.DeleteOpts{}).ExtractErr()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up volume. Please delete the volume manually: %s", s.volumeID))
	}
}
