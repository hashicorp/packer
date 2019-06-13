package uhost

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/common/retry"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type stepCopyUCloudImage struct {
	ImageDestinations []ImageDestination
	RegionId          string
	ProjectId         string
}

func (s *stepCopyUCloudImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.ImageDestinations) == 0 {
		return multistep.ActionContinue
	}

	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn
	ui := state.Get("ui").(packer.Ui)

	srcImageId := state.Get("image_id").(string)
	artifactImages := state.Get("ucloud_images").(*imageInfoSet)
	expectedImages := newImageInfoSet(nil)

	ui.Say(fmt.Sprintf("Copying image with %q...", srcImageId))
	for _, imageDestination := range s.ImageDestinations {
		if imageDestination.ProjectId == s.ProjectId && imageDestination.Region == s.RegionId {
			continue
		}

		req := conn.NewCopyCustomImageRequest()
		req.TargetProjectId = ucloud.String(imageDestination.ProjectId)
		req.TargetRegion = ucloud.String(imageDestination.Region)
		req.SourceImageId = ucloud.String(srcImageId)
		req.TargetImageName = ucloud.String(imageDestination.Name)

		resp, err := conn.CopyCustomImage(req)
		if err != nil {
			return halt(state, err, "Error on copying images")
		}

		image := imageInfo{
			Region:    imageDestination.Region,
			ProjectId: imageDestination.ProjectId,
			ImageId:   resp.TargetImageId,
		}
		expectedImages.Set(image)
		artifactImages.Set(image)

		ui.Message(fmt.Sprintf("Copying image from %s:%s:%s to %s:%s:%s)",
			s.ProjectId, s.RegionId, srcImageId, imageDestination.ProjectId, imageDestination.Region, resp.TargetImageId))
	}

	err := retry.Config{
		Tries:       200,
		ShouldRetry: func(err error) bool { return isNotCompleteError(err) },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 12 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		for _, v := range expectedImages.GetAll() {
			imageSet, err := client.describeImageByInfo(v.ProjectId, v.Region, v.ImageId)
			if err != nil {
				return err
			}

			if imageSet.State == "Available" {
				expectedImages.Remove(v.Id())
				continue
			}
		}

		if len(expectedImages.GetAll()) != 0 {
			return newNotCompleteError("copying image")
		}

		return nil
	})

	if err != nil {
		return halt(state, err, fmt.Sprintf("Error on waiting for copying image finished"))
	}

	ui.Message(fmt.Sprintf("Copy image complete"))
	return multistep.ActionContinue
}

func (s *stepCopyUCloudImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	srcImageId := state.Get("image_id").(string)
	ucloudImages := state.Get("ucloud_images").(*imageInfoSet)
	imageInfos := ucloudImages.GetAll()
	if len(imageInfos) == 0 {
		return
	} else if len(imageInfos) == 1 && imageInfos[0].ImageId == srcImageId {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn
	ui.Say(fmt.Sprintf("Deleting copied image because cancellation or error..."))

	for _, v := range imageInfos {
		if v.ImageId == srcImageId {
			continue
		}

		req := conn.NewTerminateCustomImageRequest()
		req.ProjectId = ucloud.String(v.ProjectId)
		req.Region = ucloud.String(v.Region)
		req.ImageId = ucloud.String(v.ImageId)
		_, err := conn.TerminateCustomImage(req)
		if err != nil {
			ui.Error(fmt.Sprintf("Error on deleting copied image %q", v.ImageId))
		}
	}

	ui.Message("Delete copied image complete")
}
