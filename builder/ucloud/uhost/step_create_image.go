package uhost

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/common/retry"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type stepCreateImage struct {
	image *uhost.UHostImageSet
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn
	instance := state.Get("instance").(*uhost.UHostInstanceSet)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Creating image %s", config.ImageName))

	req := conn.NewCreateCustomImageRequest()
	req.ImageName = ucloud.String(config.ImageName)
	req.ImageDescription = ucloud.String(config.ImageDescription)
	req.UHostId = ucloud.String(instance.UHostId)

	resp, err := conn.CreateCustomImage(req)
	if err != nil {
		return halt(state, err, "")
	}

	err = retry.Config{
		Tries: 200,
		ShouldRetry: func(err error) bool {
			return isExpectedStateError(err)
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 12 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		inst, err := client.DescribeImageById(resp.ImageId)
		if err != nil {
			return err
		}
		if inst == nil || inst.State != "Available" {
			return newExpectedStateError("image", resp.ImageId)
		}

		return nil
	})

	if err != nil {
		return halt(state, err, "Error on waiting for image to available")
	}

	imageSet, err := client.DescribeImageById(resp.ImageId)
	if err != nil {
		return halt(state, err, "Error on reading image")
	}

	s.image = imageSet
	state.Put("image_id", imageSet.ImageId)

	images := []imageInfo{
		{
			ImageId:   imageSet.ImageId,
			ProjectId: config.ProjectId,
			Region:    config.Region,
		},
	}

	state.Put("ucloud_images", newImageInfoSet(images))
	ui.Message(fmt.Sprintf("Create image %q complete", imageSet.ImageId))
	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting image because of cancellation or error...")
	req := conn.NewTerminateCustomImageRequest()
	req.ImageId = ucloud.String(s.image.ImageId)
	_, err := conn.TerminateCustomImage(req)
	if err != nil {
		ui.Error(fmt.Sprintf("Error on deleting image %q", s.image.ImageId))
	}
	ui.Message(fmt.Sprintf("Delete image %q complete", s.image.ImageId))
}
