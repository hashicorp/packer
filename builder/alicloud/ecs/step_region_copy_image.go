package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepRegionCopyAlicloudImage struct {
	AlicloudImageDestinationRegions []string
	AlicloudImageDestinationNames   []string
	RegionId                        string
}

func (s *stepRegionCopyAlicloudImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)

	if config.ImageEncrypted != nil {
		s.AlicloudImageDestinationRegions = append(s.AlicloudImageDestinationRegions, s.RegionId)
		s.AlicloudImageDestinationNames = append(s.AlicloudImageDestinationNames, config.AlicloudImageName)
	}

	if len(s.AlicloudImageDestinationRegions) == 0 {
		return multistep.ActionContinue
	}

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)

	srcImageId := state.Get("alicloudimage").(string)
	alicloudImages := state.Get("alicloudimages").(map[string]string)
	numberOfName := len(s.AlicloudImageDestinationNames)

	ui.Say(fmt.Sprintf("Coping image %s from %s...", srcImageId, s.RegionId))
	for index, destinationRegion := range s.AlicloudImageDestinationRegions {
		if destinationRegion == s.RegionId && config.ImageEncrypted == nil {
			continue
		}

		ecsImageName := ""
		if numberOfName > 0 && index < numberOfName {
			ecsImageName = s.AlicloudImageDestinationNames[index]
		}

		copyImageRequest := ecs.CreateCopyImageRequest()
		copyImageRequest.RegionId = s.RegionId
		copyImageRequest.ImageId = srcImageId
		copyImageRequest.DestinationRegionId = destinationRegion
		copyImageRequest.DestinationImageName = ecsImageName
		if config.ImageEncrypted != nil {
			copyImageRequest.Encrypted = requests.NewBoolean(*config.ImageEncrypted)
		}

		imageResponse, err := client.CopyImage(copyImageRequest)
		if err != nil {
			return halt(state, err, "Error copying images")
		}

		alicloudImages[destinationRegion] = imageResponse.ImageId
		ui.Message(fmt.Sprintf("Copy image from %s(%s) to %s(%s)", s.RegionId, srcImageId, destinationRegion, imageResponse.ImageId))
	}

	if config.ImageEncrypted != nil {
		if _, err := client.WaitForImageStatus(s.RegionId, alicloudImages[s.RegionId], ImageStatusAvailable, time.Duration(ALICLOUD_DEFAULT_LONG_TIMEOUT)*time.Second); err != nil {
			return halt(state, err, fmt.Sprintf("Timeout waiting image %s finish copying", alicloudImages[s.RegionId]))
		}
	}

	return multistep.ActionContinue
}

func (s *stepRegionCopyAlicloudImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("Stopping copy image because cancellation or error..."))

	client := state.Get("client").(*ClientWrapper)
	alicloudImages := state.Get("alicloudimages").(map[string]string)
	srcImageId := state.Get("alicloudimage").(string)

	for copiedRegionId, copiedImageId := range alicloudImages {
		if copiedImageId == srcImageId {
			continue
		}

		cancelCopyImageRequest := ecs.CreateCancelCopyImageRequest()
		cancelCopyImageRequest.RegionId = copiedRegionId
		cancelCopyImageRequest.ImageId = copiedImageId
		if _, err := client.CancelCopyImage(cancelCopyImageRequest); err != nil {

			ui.Error(fmt.Sprintf("Error cancelling copy image: %v", err))
		}
	}
}
