package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepDeleteAlicloudImageSnapshots struct {
	AlicloudImageForceDelete          bool
	AlicloudImageForceDeleteSnapshots bool
	AlicloudImageName                 string
	AlicloudImageDestinationRegions   []string
	AlicloudImageDestinationNames     []string
}

func (s *stepDeleteAlicloudImageSnapshots) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	images, _, err := client.DescribeImages(&ecs.DescribeImagesArgs{
		RegionId:  common.Region(region),
		ImageName: imageName,
	})
	if len(images) < 1 {
		return nil
	}

	ui.Say(fmt.Sprintf("Deleting duplicated image and snapshot in %s: %s", region, imageName))

	for _, image := range images {
		if image.ImageOwnerAlias != string(ecs.ImageOwnerSelf) {
			log.Printf("You can not delete non-customized images: %s ", image.ImageId)
			continue
		}

		err = client.DeleteImage(common.Region(region), image.ImageId)
		if err != nil {
			err := fmt.Errorf("Failed to delete image: %s", err)
			return err
		}

		if s.AlicloudImageForceDeleteSnapshots {
			for _, diskDevice := range image.DiskDeviceMappings.DiskDeviceMapping {
				if err := client.DeleteSnapshot(diskDevice.SnapshotId); err != nil {
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
