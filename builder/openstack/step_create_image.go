package openstack

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumeactions"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreateImage struct {
	UseBlockStorageVolume bool
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	server := state.Get("server").(*servers.Server)
	ui := state.Get("ui").(packersdk.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// We need the v2 image client
	imageClient, err := config.imageV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing image service client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Create the image.
	// Image source depends on the type of the Compute instance. It can be
	// Block Storage service volume or regular Compute service local volume.
	ui.Say(fmt.Sprintf("Creating the image: %s", config.ImageName))
	var imageId string
	var blockStorageClient *gophercloud.ServiceClient
	if s.UseBlockStorageVolume {
		// We need the v3 block storage client.
		blockStorageClient, err = config.blockStorageV3Client()
		if err != nil {
			err = fmt.Errorf("Error initializing block storage client: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		}
		volume := state.Get("volume_id").(string)

		// set ImageMetadata before uploading to glance so the new image captured the desired values
		if len(config.ImageMetadata) > 0 {
			err = volumeactions.SetImageMetadata(blockStorageClient, volume, volumeactions.ImageMetadataOpts{
				Metadata: config.ImageMetadata,
			}).ExtractErr()
			if err != nil {
				err := fmt.Errorf("Error setting image metadata: %s", err)
				ui.Error(err.Error())
			}
		}

		image, err := volumeactions.UploadImage(blockStorageClient, volume, volumeactions.UploadImageOpts{
			DiskFormat: config.ImageDiskFormat,
			ImageName:  config.ImageName,
		}).Extract()
		if err != nil {
			err := fmt.Errorf("Error creating image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		imageId = image.ImageID
	} else {
		imageId, err = servers.CreateImage(computeClient, server.ID, servers.CreateImageOpts{
			Name:     config.ImageName,
			Metadata: config.ImageMetadata,
		}).ExtractImageID()
		if err != nil {
			err := fmt.Errorf("Error creating image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Set the Image ID in the state
	ui.Message(fmt.Sprintf("Image: %s", imageId))
	state.Put("image", imageId)

	// Wait for the image to become ready
	ui.Say(fmt.Sprintf("Waiting for image %s (image id: %s) to become ready...", config.ImageName, imageId))
	if err := WaitForImage(ctx, imageClient, imageId); err != nil {
		err := fmt.Errorf("Error waiting for image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(multistep.StateBag) {
	// No cleanup...
}

// WaitForImage waits for the given Image ID to become ready.
func WaitForImage(ctx context.Context, client *gophercloud.ServiceClient, imageId string) error {
	maxNumErrors := 10
	numErrors := 0

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		image, err := images.Get(client, imageId).Extract()
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

		if image.Status == "active" {
			return nil
		}

		log.Printf("Waiting for image creation status: %s", image.Status)
		time.Sleep(2 * time.Second)
	}
}
