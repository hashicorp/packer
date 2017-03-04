package ecs

import (
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepDeleteAlicloudImageSnapshots struct {
	AlicloudImageForceDetele          bool
	AlicloudImageForceDeteleSnapshots bool
	AlicloudImageName                 string
}

func (s *stepDeleteAlicloudImageSnapshots) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	ui.Say("Start delete alicloud image snapshots")
	// Check for force delete
	if s.AlicloudImageForceDetele {
		images, _, err := client.DescribeImages(&ecs.DescribeImagesArgs{
			RegionId:  common.Region(config.AlicloudRegion),
			ImageName: s.AlicloudImageName,
		})
		if len(images) < 1 {
			return multistep.ActionContinue
		}
		for _, image := range images {
			if image.ImageOwnerAlias != string(ecs.ImageOwnerSelf) {
				log.Printf("Only can delete the instance based on customized images %s ", image.ImageId)
				continue
			}
			err = client.DeleteImage(common.Region(config.AlicloudRegion), image.ImageId)
			if err != nil {
				err := fmt.Errorf("Delete alicloud image failed: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			if s.AlicloudImageForceDeteleSnapshots {
				for _, diskDevice := range image.DiskDeviceMappings.DiskDeviceMapping {
					if err := client.DeleteSnapshot(diskDevice.SnapshotId); err != nil {
						err := fmt.Errorf("Deleting ECS snapshot failed: %s", err)
						state.Put("error", err)
						ui.Error(err.Error())
						return multistep.ActionHalt
					}
				}
			}
		}

	}

	return multistep.ActionContinue
}

func (s *stepDeleteAlicloudImageSnapshots) Cleanup(state multistep.StateBag) {
}
