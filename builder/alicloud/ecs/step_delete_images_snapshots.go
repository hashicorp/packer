package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepDeleteAlicloudImageSnapshots struct {
	AlicloudImageForceDelete          bool
	AlicloudImageForceDeleteSnapshots bool
	AlicloudImageName                 string
	AlicloudImageDestinationRegions   []string
	AlicloudImageDestinationNames     []string
}

func (s *stepDeleteAlicloudImageSnapshots) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)

	// Check for force delete
	if s.AlicloudImageForceDelete {
		err := s.deleteImageAndSnapshots(state, s.AlicloudImageName, config.AlicloudRegion)
		if err != nil {
			return halt(state, err, "")
		}

		numberOfName := len(s.AlicloudImageDestinationNames)
		if numberOfName == 0 {
			return multistep.ActionContinue
		}

		for index, destinationRegion := range s.AlicloudImageDestinationRegions {
			if destinationRegion == config.AlicloudRegion {
				continue
			}

			if index < numberOfName {
				err = s.deleteImageAndSnapshots(state, s.AlicloudImageDestinationNames[index], destinationRegion)
				if err != nil {
					return halt(state, err, "")
				}
			} else {
				break
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepDeleteAlicloudImageSnapshots) deleteImageAndSnapshots(state multistep.StateBag, imageName string, region string) error {
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.RegionId = region
	describeImagesRequest.ImageName = imageName
	describeImagesRequest.Status = ImageStatusQueried
	imageResponse, _ := client.DescribeImages(describeImagesRequest)
	images := imageResponse.Images.Image
	if len(images) < 1 {
		return nil
	}

	ui.Say(fmt.Sprintf("Deleting duplicated image and snapshot in %s: %s", region, imageName))

	for _, image := range images {
		if image.ImageOwnerAlias != ImageOwnerSelf {
			log.Printf("You can not delete non-customized images: %s ", image.ImageId)
			continue
		}

		deleteImageRequest := ecs.CreateDeleteImageRequest()
		deleteImageRequest.RegionId = region
		deleteImageRequest.ImageId = image.ImageId
		if _, err := client.DeleteImage(deleteImageRequest); err != nil {
			err := fmt.Errorf("Failed to delete image: %s", err)
			return err
		}

		if s.AlicloudImageForceDeleteSnapshots {
			for _, diskDevice := range image.DiskDeviceMappings.DiskDeviceMapping {
				deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
				deleteSnapshotRequest.SnapshotId = diskDevice.SnapshotId
				if _, err := client.DeleteSnapshot(deleteSnapshotRequest); err != nil {
					err := fmt.Errorf("Deleting ECS snapshot failed: %s", err)
					return err
				}
			}
		}
	}

	return nil
}

func (s *stepDeleteAlicloudImageSnapshots) Cleanup(state multistep.StateBag) {
}
