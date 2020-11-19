package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCheckAlicloudSourceImage struct {
	SourceECSImageId string
}

func (s *stepCheckAlicloudSourceImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.RegionId = config.AlicloudRegion
	describeImagesRequest.ImageId = config.AlicloudSourceImage
	if config.AlicloudSkipImageValidation {
		describeImagesRequest.ShowExpired = "true"
	}
	imagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return halt(state, err, "Error querying alicloud image")
	}

	images := imagesResponse.Images.Image

	// Describe marketplace image
	describeImagesRequest.ImageOwnerAlias = "marketplace"
	marketImagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return halt(state, err, "Error querying alicloud marketplace image")
	}

	marketImages := marketImagesResponse.Images.Image
	if len(marketImages) > 0 {
		images = append(images, marketImages...)
	}

	if len(images) == 0 {
		err := fmt.Errorf("No alicloud image was found matching filters: %v", config.AlicloudSourceImage)
		return halt(state, err, "")
	}

	ui.Message(fmt.Sprintf("Found image ID: %s", images[0].ImageId))

	state.Put("source_image", &images[0])
	return multistep.ActionContinue
}

func (s *stepCheckAlicloudSourceImage) Cleanup(multistep.StateBag) {}
