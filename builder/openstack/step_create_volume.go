package openstack

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateVolume struct {
	UseBlockStorageVolume  bool
	SourceImage            string
	VolumeName             string
	VolumeType             string
	VolumeAvailabilityZone string
	volumeID               string
	doCleanup              bool
}

func (s *StepCreateVolume) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Proceed only if block storage volume is required.
	if !s.UseBlockStorageVolume {
		return multistep.ActionContinue
	}

	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We will need Block Storage and Image services clients.
	blockStorageClient, err := config.blockStorageV3Client()
	if err != nil {
		err = fmt.Errorf("Error initializing block storage client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}
	imageClient, err := config.imageV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing image client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Get needed volume size from the source image.
	volumeSize, err := GetVolumeSize(imageClient, s.SourceImage)
	if err != nil {
		err := fmt.Errorf("Error creating volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Creating volume...")
	volumeOpts := volumes.CreateOpts{
		Size:             volumeSize,
		VolumeType:       s.VolumeType,
		AvailabilityZone: s.VolumeAvailabilityZone,
		Name:             s.VolumeName,
		ImageID:          s.SourceImage,
	}

	volume, err := volumes.Create(blockStorageClient, volumeOpts).Extract()
	if err != nil {
		err := fmt.Errorf("Error creating volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the volume to become ready.
	ui.Say(fmt.Sprintf("Waiting for volume %s (volume id: %s) to become ready...", config.VolumeName, volume.ID))
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

// GetVolumeSize returns volume size in gigabytes based on the image min disk
// value if it's not empty.
// Or it calculates needed gigabytes size from the image bytes size.
func GetVolumeSize(imageClient *gophercloud.ServiceClient, imageID string) (int, error) {
	sourceImage, err := images.Get(imageClient, imageID).Extract()
	if err != nil {
		return 0, err
	}

	if sourceImage.MinDiskGigabytes != 0 {
		return sourceImage.MinDiskGigabytes, nil
	}

	volumeSizeMB := sourceImage.SizeBytes / 1024 / 1024
	volumeSizeGB := int(sourceImage.SizeBytes / 1024 / 1024 / 1024)

	// Increment gigabytes size if the initial size can't be divided without
	// remainder.
	if volumeSizeMB%1024 > 0 {
		volumeSizeGB++
	}

	return volumeSizeGB, nil
}

// WaitForVolume waits for the given volume to become ready.
func WaitForVolume(blockStorageClient *gophercloud.ServiceClient, volumeID string) error {
	maxNumErrors := 10
	numErrors := 0

	for {
		volume, err := volumes.Get(blockStorageClient, volumeID).Extract()
		if err != nil {
			errCode, ok := err.(*gophercloud.ErrUnexpectedResponseCode)
			if ok && (errCode.Actual == 500 || errCode.Actual == 404) {
				numErrors++
				if numErrors >= maxNumErrors {
					log.Printf("[ERROR] Maximum number of errors (%d) reached; failing with: %s", numErrors, err)
					return err
				}
				log.Printf("[ERROR] %d error received, will ignore and retry: %s", errCode.Actual, err)
				time.Sleep(2 * time.Second)
				continue
			}

			return err
		}

		if volume.Status == "available" {
			return nil
		}

		log.Printf("Waiting for volume creation status: %s", volume.Status)
		time.Sleep(2 * time.Second)
	}
}

func (s *StepCreateVolume) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}

	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	blockStorageClient, err := config.blockStorageV3Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up volume. Please delete the volume manually: %s", s.volumeID))
		return
	}

	ui.Say(fmt.Sprintf("Deleting volume: %s ...", s.volumeID))
	err = volumes.Delete(blockStorageClient, s.volumeID).ExtractErr()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up volume. Please delete the volume manually: %s", s.volumeID))
	}
}
