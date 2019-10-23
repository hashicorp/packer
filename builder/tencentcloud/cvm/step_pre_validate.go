package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepPreValidate struct {
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("cvm_client").(*cvm.Client)

	Say(state, config.ImageName, "Trying to check image name")

	req := cvm.NewDescribeImagesRequest()
	req.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("image-name"),
			Values: []*string{&config.ImageName},
		},
	}
	var resp *cvm.DescribeImagesResponse
	err := Retry(ctx, func(ctx context.Context) error {
		var err error
		resp, err = client.DescribeImages(req)
		return err
	})
	if err != nil {
		return Halt(state, err, "Failed to get images info")
	}

	if *resp.Response.TotalCount > 0 {
		return Halt(state, fmt.Errorf("Image name %s has exists", config.ImageName), "")
	}

	Message(state, "useable", "Image name")

	return multistep.ActionContinue
}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {}
