package ecs

import (
	"context"
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateAlicloudImage struct {
	AlicloudImageIgnoreDataDisks bool
	WaitSnapshotReadyTimeout     int
	image                        *ecs.ImageType
}

func (s *stepCreateAlicloudImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	// Create the alicloud image
	ui.Say(fmt.Sprintf("Creating image: %s", config.AlicloudImageName))
	var imageId string
	var err error

	if s.AlicloudImageIgnoreDataDisks {
		snapshotId := state.Get("alicloudsnapshot").(string)
		imageId, err = client.CreateImage(&ecs.CreateImageArgs{
			RegionId:     common.Region(config.AlicloudRegion),
			SnapshotId:   snapshotId,
			ImageName:    config.AlicloudImageName,
			ImageVersion: config.AlicloudImageVersion,
			Description:  config.AlicloudImageDescription})
	} else {
		instance := state.Get("instance").(*ecs.InstanceAttributesType)
		imageId, err = client.CreateImage(&ecs.CreateImageArgs{
			RegionId:     common.Region(config.AlicloudRegion),
			InstanceId:   instance.InstanceId,
			ImageName:    config.AlicloudImageName,
			ImageVersion: config.AlicloudImageVersion,
			Description:  config.AlicloudImageDescription})
	}

	if err != nil {
		return halt(state, err, "Error creating image")
	}
	err = client.WaitForImageReady(common.Region(config.AlicloudRegion), imageId, s.WaitSnapshotReadyTimeout)
	if err != nil {
		return halt(state, err, "Timeout waiting for image to be created")
	}

	images, _, err := client.DescribeImages(&ecs.DescribeImagesArgs{
		RegionId: common.Region(config.AlicloudRegion),
		ImageId:  imageId})
	if err != nil {
		return halt(state, err, "Error querying created imaged")
	}

	if len(images) == 0 {
		return halt(state, err, "Unable to find created image")
	}

	s.image = &images[0]

	var snapshotIds = []string{}
	for _, device := range images[0].DiskDeviceMappings.DiskDeviceMapping {
		snapshotIds = append(snapshotIds, device.SnapshotId)
	}

	state.Put("alicloudimage", imageId)
	state.Put("alicloudsnapshots", snapshotIds)

	alicloudImages := make(map[string]string)
	alicloudImages[config.AlicloudRegion] = images[0].ImageId
	state.Put("alicloudimages", alicloudImages)

	return multistep.ActionContinue
}

func (s *stepCreateAlicloudImage) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	ui.Say("Deleting the image because of cancellation or error...")
	if err := client.DeleteImage(common.Region(config.AlicloudRegion), s.image.ImageId); err != nil {
		ui.Error(fmt.Sprintf("Error deleting image, it may still be around: %s", err))
		return
	}
}
