package ecs

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCheckAlicloudSourceImage struct {
	SourceECSImageId string
}

func (s *stepCheckAlicloudSourceImage) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	images, _, err := client.DescribeImages(&ecs.DescribeImagesArgs{RegionId: common.Region(config.AlicloudRegion),
		ImageId: config.AlicloudSourceImage})
	if err != nil {
		err := fmt.Errorf("Error querying alicloud image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(images) == 0 {
		err := fmt.Errorf("No alicloud image was found matching filters: %v", config.AlicloudSourceImage)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Found image ID: %s", images[0].ImageId))

	state.Put("source_image", &images[0])
	return multistep.ActionContinue
}

func (s *stepCheckAlicloudSourceImage) Cleanup(multistep.StateBag) {}
