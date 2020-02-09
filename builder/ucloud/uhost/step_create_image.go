package uhost

import (
	"context"
	"fmt"
	"time"

	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/common/retry"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type stepCreateImage struct {
	image *uhost.UHostImageSet
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ucloudcommon.UCloudClient)
	conn := client.UHostConn
	instance := state.Get("instance").(*uhost.UHostInstanceSet)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Creating image %s...", config.ImageName))

	req := conn.NewCreateCustomImageRequest()
	req.ImageName = ucloud.String(config.ImageName)
	req.ImageDescription = ucloud.String(config.ImageDescription)
	req.UHostId = ucloud.String(instance.UHostId)

	resp, err := conn.CreateCustomImage(req)
	if err != nil {
		return ucloudcommon.Halt(state, err, "Error on creating image")
	}
	ui.Message(fmt.Sprintf("Waiting for the created image %q to become available...", resp.ImageId))

	err = retry.Config{
		StartTimeout: time.Duration(config.WaitImageReadyTimeout) * time.Second,
		ShouldRetry: func(err error) bool {
			return ucloudcommon.IsExpectedStateError(err)
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 12 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		inst, err := client.DescribeImageById(resp.ImageId)
		if err != nil {
			return err
		}
		if inst == nil || inst.State != ucloudcommon.ImageStateAvailable {
			return ucloudcommon.NewExpectedStateError("image", resp.ImageId)
		}

		return nil
	})

	if err != nil {
		return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on waiting for image %q to become available", resp.ImageId))
	}

	imageSet, err := client.DescribeImageById(resp.ImageId)
	if err != nil {
		return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on reading image when creating %q", resp.ImageId))
	}

	s.image = imageSet
	state.Put("image_id", imageSet.ImageId)

	images := []ucloudcommon.ImageInfo{
		{
			ImageId:   imageSet.ImageId,
			ProjectId: config.ProjectId,
			Region:    config.Region,
		},
	}

	state.Put("ucloud_images", ucloudcommon.NewImageInfoSet(images))
	ui.Message(fmt.Sprintf("Creating image %q complete", imageSet.ImageId))
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

	client := state.Get("client").(*ucloudcommon.UCloudClient)
	conn := client.UHostConn
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting image because of cancellation or error...")
	req := conn.NewTerminateCustomImageRequest()
	req.ImageId = ucloud.String(s.image.ImageId)
	_, err := conn.TerminateCustomImage(req)
	if err != nil {
		ui.Error(fmt.Sprintf("Error on deleting image %q", s.image.ImageId))
	}
	ui.Message(fmt.Sprintf("Deleting image %q complete", s.image.ImageId))
}
