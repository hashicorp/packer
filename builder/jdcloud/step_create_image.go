package jdcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
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

	s.InstanceSpecConfig.ArtifactId = resp.Result.ImageId
	if err := ImageStatusWaiter(s.InstanceSpecConfig.ArtifactId); err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

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

}

// Delete created instance image on error
func (s *stepCreateJDCloudImage) Cleanup(state multistep.StateBag) {

	if s.InstanceSpecConfig.ArtifactId != "" {

		req := apis.NewDeleteImageRequest(Region, s.InstanceSpecConfig.ArtifactId)

		_ = Retry(time.Minute, func() *RetryError {
			_, err := VmClient.DeleteImage(req)
			if err == nil {
				return nil
			}
			if connectionError(err) {
				return RetryableError(err)
			} else {
				return NonRetryableError(err)
			}
		})
	}

}
