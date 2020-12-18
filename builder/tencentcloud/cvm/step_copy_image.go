package cvm

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCopyImage struct {
	DesinationRegions []string
	SourceRegion      string
}

func (s *stepCopyImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.DesinationRegions) == 0 || (len(s.DesinationRegions) == 1 && s.DesinationRegions[0] == s.SourceRegion) {
		return multistep.ActionContinue
	}

	config := state.Get("config").(*Config)
	client := state.Get("cvm_client").(*cvm.Client)

	imageId := state.Get("image").(*cvm.Image).ImageId

	Say(state, strings.Join(s.DesinationRegions, ","), "Trying to copy image to")

	req := cvm.NewSyncImagesRequest()
	req.ImageIds = []*string{imageId}
	copyRegions := make([]*string, 0, len(s.DesinationRegions))
	for _, region := range s.DesinationRegions {
		if region != s.SourceRegion {
			copyRegions = append(copyRegions, common.StringPtr(region))
		}
	}
	req.DestinationRegions = copyRegions

	err := Retry(ctx, func(ctx context.Context) error {
		_, e := client.SyncImages(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to copy image")
	}

	Message(state, "Waiting for image ready", "")
	tencentCloudImages := state.Get("tencentcloudimages").(map[string]string)

	for _, region := range req.DestinationRegions {
		rc, err := NewCvmClient(config.SecretId, config.SecretKey, *region)
		if err != nil {
			return Halt(state, err, "Failed to init client")
		}

		err = WaitForImageReady(ctx, rc, config.ImageName, "NORMAL", 1800)
		if err != nil {
			return Halt(state, err, "Failed to wait for image ready")
		}

		image, err := GetImageByName(ctx, rc, config.ImageName)
		if err != nil {
			return Halt(state, err, "Failed to get image")
		}

		if image == nil {
			return Halt(state, err, "Failed to wait for image ready")
		}

		tencentCloudImages[*region] = *image.ImageId
		Message(state, fmt.Sprintf("Copy image from %s(%s) to %s(%s)", s.SourceRegion, *imageId, *region, *image.ImageId), "")
	}

	state.Put("tencentcloudimages", tencentCloudImages)
	Message(state, "Image copied", "")

	return multistep.ActionContinue
}

func (s *stepCopyImage) Cleanup(state multistep.StateBag) {}
