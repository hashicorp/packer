package ecs

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type setpRegionCopyAlicloudImage struct {
	AlicloudImageDestinationRegions []string
	AlicloudImageDestinationNames   []string
	RegionId                        string
}

func (s *setpRegionCopyAlicloudImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.AlicloudImageDestinationRegions) == 0 {
		return multistep.ActionContinue
	}
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	imageId := state.Get("alicloudimage").(string)
	alicloudImages := state.Get("alicloudimages").(map[string]string)
	region := common.Region(s.RegionId)

	numberOfName := len(s.AlicloudImageDestinationNames)
	for index, destinationRegion := range s.AlicloudImageDestinationRegions {
		if destinationRegion == s.RegionId {
			continue
		}
		ecsImageName := ""
		if numberOfName > 0 && index < numberOfName {
			ecsImageName = s.AlicloudImageDestinationNames[index]
		}
		imageId, err := client.CopyImage(
			&ecs.CopyImageArgs{
				RegionId:             region,
				ImageId:              imageId,
				DestinationRegionId:  common.Region(destinationRegion),
				DestinationImageName: ecsImageName,
			})
		if err != nil {
			state.Put("error", err)
			ui.Say(fmt.Sprintf("Error copying images: %s", err))
			return multistep.ActionHalt
		}
		alicloudImages[destinationRegion] = imageId
	}
	return multistep.ActionContinue
}

func (s *setpRegionCopyAlicloudImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if cancelled || halted {
		ui := state.Get("ui").(packer.Ui)
		client := state.Get("client").(*ecs.Client)
		alicloudImages := state.Get("alicloudimages").(map[string]string)
		ui.Say(fmt.Sprintf("Stopping copy image because cancellation or error..."))
		for copiedRegionId, copiedImageId := range alicloudImages {
			if copiedRegionId == s.RegionId {
				continue
			}
			if err := client.CancelCopyImage(common.Region(copiedRegionId), copiedImageId); err != nil {
				ui.Say(fmt.Sprintf("Error cancelling copy image: %v", err))
			}
		}
	}
}
