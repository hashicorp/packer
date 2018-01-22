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
	image *ecs.ImageType
}

func (s *stepCreateAlicloudImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	// Create the alicloud image
	ui.Say(fmt.Sprintf("Creating image: %s", config.AlicloudImageName))
	var imageId string
	var err error

	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	imageId, err = client.CreateImage(&ecs.CreateImageArgs{
		RegionId:     common.Region(config.AlicloudRegion),
		InstanceId:   instance.InstanceId,
		ImageName:    config.AlicloudImageName,
		ImageVersion: config.AlicloudImageVersion,
		Description:  config.AlicloudImageDescription})

	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	err = client.WaitForImageReady(common.Region(config.AlicloudRegion),
		imageId, ALICLOUD_DEFAULT_LONG_TIMEOUT)
	if err != nil {
		err := fmt.Errorf("Timeout waiting for image to be created: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	images, _, err := client.DescribeImages(&ecs.DescribeImagesArgs{
		RegionId: common.Region(config.AlicloudRegion),
		ImageId:  imageId})
	if err != nil {
		err := fmt.Errorf("Error querying created image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(images) == 0 {
		err := fmt.Errorf("Unable to find created image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.image = &images[0]

	state.Put("alicloudimage", imageId)
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
	config := state.Get("config").(Config)

	ui.Say("Deleting the image because of cancellation or error...")
	if err := client.DeleteImage(common.Region(config.AlicloudRegion), s.image.ImageId); err != nil {
		ui.Error(fmt.Sprintf("Error deleting image, it may still be around: %s", err))
		return
	}
}
