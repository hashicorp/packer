package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCreateImage struct {
	imageId string
}

func (s *stepCreateImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("cvm_client").(*cvm.Client)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*cvm.Instance)

	ui.Say(fmt.Sprintf("Creating image %s", config.ImageName))

	req := cvm.NewCreateImageRequest()
	req.ImageName = &config.ImageName
	req.ImageDescription = &config.ImageDescription
	req.InstanceId = instance.InstanceId

	True := "True"
	False := "False"
	if config.ForcePoweroff {
		req.ForcePoweroff = &True
	} else {
		req.ForcePoweroff = &False
	}

	if config.Reboot {
		req.Reboot = &True
	} else {
		req.Reboot = &False
	}

	if config.Sysprep {
		req.Sysprep = &True
	} else {
		req.Sysprep = &False
	}

	_, err := client.CreateImage(req)
	if err != nil {
		err := fmt.Errorf("create image failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = WaitForImageReady(client, config.ImageName, "NORMAL", 3600)
	if err != nil {
		err := fmt.Errorf("create image failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	describeReq := cvm.NewDescribeImagesRequest()
	FILTER_IMAGE_NAME := "image-name"
	describeReq.Filters = []*cvm.Filter{
		{
			Name:   &FILTER_IMAGE_NAME,
			Values: []*string{&config.ImageName},
		},
	}
	describeResp, err := client.DescribeImages(describeReq)
	if err != nil {
		err := fmt.Errorf("wait image ready failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if *describeResp.Response.TotalCount == 0 {
		err := fmt.Errorf("create image(%s) failed", config.ImageName)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.imageId = *describeResp.Response.ImageSet[0].ImageId
	state.Put("image", describeResp.Response.ImageSet[0])

	tencentCloudImages := make(map[string]string)
	tencentCloudImages[config.Region] = s.imageId
	state.Put("tencentcloudimages", tencentCloudImages)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	if s.imageId == "" {
		return
	}
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("cvm_client").(*cvm.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Delete image because of cancellation or error...")
	req := cvm.NewDeleteImagesRequest()
	req.ImageIds = []*string{&s.imageId}
	_, err := client.DeleteImages(req)
	if err != nil {
		ui.Error(fmt.Sprintf("delete image(%s) failed", s.imageId))
	}
}
