package jdcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
	"time"
)

type stepCreateJDCloudImage struct {
	InstanceSpecConfig *JDCloudInstanceSpecConfig
}

func (s *stepCreateJDCloudImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating images")

	req := apis.NewCreateImageRequest(Region, s.InstanceSpecConfig.InstanceId, s.InstanceSpecConfig.ImageName, "")
	resp, err := VmClient.CreateImage(req)
	if err != nil || resp.Error.Code != FINE {
		ui.Error(fmt.Sprintf("[ERROR] Creating image: Error-%v ,Resp:%v", err, resp))
		return multistep.ActionHalt
	}

	imageId := resp.Result.ImageId
	if err := ImageStatusWaiter(imageId); err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.InstanceSpecConfig.ArtifactId = imageId
	return multistep.ActionContinue
}

func ImageStatusWaiter(imageId string) error {
	req := apis.NewDescribeImageRequest(Region, imageId)

	return Retry(5*time.Minute, func() *RetryError {
		resp, err := VmClient.DescribeImage(req)
		if err == nil && resp.Result.Image.Status == READY {
			return nil
		}
		if connectionError(err) {
			return RetryableError(err)
		} else {
			return NonRetryableError(err)
		}
	})
	return nil
}

func (s *stepCreateJDCloudImage) Cleanup(state multistep.StateBag) {
	return
}
