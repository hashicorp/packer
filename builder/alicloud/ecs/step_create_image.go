package ecs

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"time"
)

type stepCreateAlicloudImage struct {
	AlicloudImageIgnoreDataDisks bool
	WaitSnapshotReadyTimeout     int
	image                        *ecs.Image
}

var createImageRetryErrors = []string{
	"IdempotentProcessing",
}

func (s *stepCreateAlicloudImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)

	// Create the alicloud image
	ui.Say(fmt.Sprintf("Creating image: %s", config.AlicloudImageName))

	createImageRequest := s.buildCreateImageRequest(state)
	createImageResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.CreateImage(createImageRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createImageRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Error creating image")
	}

	imageId := createImageResponse.(*ecs.CreateImageResponse).ImageId

	_, err = client.WaitForImageStatus(config.AlicloudRegion, imageId, ImageStatusAvailable, time.Duration(s.WaitSnapshotReadyTimeout)*time.Second)
	if err != nil {
		return halt(state, err, "Timeout waiting for image to be created")
	}

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.ImageId = imageId
	describeImagesRequest.RegionId = config.AlicloudRegion
	imagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return halt(state, err, "")
	}

	images := imagesResponse.Images.Image
	if len(images) == 0 {
		return halt(state, err, "Unable to find created image")
	}

	s.image = &images[0]

	var snapshotIds []string
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

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	ui.Say("Deleting the image because of cancellation or error...")

	deleteImageRequest := ecs.CreateDeleteImageRequest()
	deleteImageRequest.RegionId = config.AlicloudRegion
	deleteImageRequest.ImageId = s.image.ImageId
	if _, err := client.DeleteImage(deleteImageRequest); err != nil {
		ui.Error(fmt.Sprintf("Error deleting image, it may still be around: %s", err))
		return
	}
}

func (s *stepCreateAlicloudImage) buildCreateImageRequest(state multistep.StateBag) *ecs.CreateImageRequest {
	config := state.Get("config").(*Config)

	request := ecs.CreateCreateImageRequest()
	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = config.AlicloudRegion
	request.ImageName = config.AlicloudImageName
	request.ImageVersion = config.AlicloudImageVersion
	request.Description = config.AlicloudImageDescription

	if s.AlicloudImageIgnoreDataDisks {
		snapshotId := state.Get("alicloudsnapshot").(string)
		request.SnapshotId = snapshotId
	} else {
		instance := state.Get("instance").(*ecs.Instance)
		request.InstanceId = instance.InstanceId
	}

	return request
}
