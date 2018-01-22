package ecs

import (
	"context"
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepPreValidate struct {
	AlicloudDestImageName string
	ForceDelete           bool
}

func (s *stepPreValidate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if s.ForceDelete {
		ui.Say("Force delete flag found, skipping prevalidating image name.")
		return multistep.ActionContinue
	}

	client := state.Get("client").(*ecs.Client)
	config := state.Get("config").(Config)
	ui.Say("Prevalidating image name...")
	images, _, err := client.DescribeImages(&ecs.DescribeImagesArgs{
		ImageName: s.AlicloudDestImageName,
		RegionId:  common.Region(config.AlicloudRegion)})

	if err != nil {
		err := fmt.Errorf("Error querying alicloud image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(images) > 0 {
		err := fmt.Errorf("Error: name conflicts with an existing alicloud image: %s", images[0].ImageId)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {}
