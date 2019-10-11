package uhost

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/common/retry"
	"strings"
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
	ui.Say(fmt.Sprintf("Copying images from %q...", srcImageId))
	for _, v := range s.ImageDestinations {
		if v.ProjectId == s.ProjectId && v.Region == s.RegionId {
			continue
		}

		req := conn.NewCopyCustomImageRequest()
		req.TargetProjectId = ucloud.String(v.ProjectId)
		req.TargetRegion = ucloud.String(v.Region)
		req.SourceImageId = ucloud.String(srcImageId)
		req.TargetImageName = ucloud.String(v.Name)
		req.TargetImageDescription = ucloud.String(v.Description)

		resp, err := conn.CopyCustomImage(req)
		if err != nil {
			return halt(state, err, fmt.Sprintf("Error on copying image %q to %s:%s", srcImageId, v.ProjectId, v.Region))
		}

		image := imageInfo{
			Region:    v.Region,
			ProjectId: v.ProjectId,
			ImageId:   resp.TargetImageId,
		}
		expectedImages.Set(image)
		artifactImages.Set(image)

		ui.Message(fmt.Sprintf("Copying image from %s:%s:%s to %s:%s:%s)",
			s.ProjectId, s.RegionId, srcImageId, v.ProjectId, v.Region, resp.TargetImageId))
	}
	ui.Message("Waiting for the copied images to become available...")

	err := retry.Config{
		Tries:       200,
		ShouldRetry: func(err error) bool { return isNotCompleteError(err) },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 12 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		for _, v := range expectedImages.GetAll() {
			imageSet, err := client.describeImageByInfo(v.ProjectId, v.Region, v.ImageId)
			if err != nil {
				return fmt.Errorf("reading %s:%s:%s failed, %s", v.ProjectId, v.Region, v.ImageId, err)
			}

			if imageSet.State == imageStateAvailable {
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
		var s []string
		for _, v := range expectedImages.GetAll() {
			s = append(s, fmt.Sprintf("%s:%s:%s", v.ProjectId, v.Region, v.ImageId))
		}

		return halt(state, err, fmt.Sprintf("Error on waiting for copying images %q to become available", strings.Join(s, ",")))
	}

	ui.Message(fmt.Sprintf("Copying image complete"))
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
	ui.Say(fmt.Sprintf("Deleting copied image because of cancellation or error..."))

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

	ui.Message("Deleting copied image complete")
}
