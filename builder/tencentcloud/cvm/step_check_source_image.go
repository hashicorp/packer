package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCheckSourceImage struct {
	sourceImageId string
}

func (s *stepCheckSourceImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("cvm_client").(*cvm.Client)

	Say(state, config.SourceImageId, "Trying to check source image")

	req := cvm.NewDescribeImagesRequest()
	req.ImageIds = []*string{&config.SourceImageId}
	req.InstanceType = &config.InstanceType
	var resp *cvm.DescribeImagesResponse
	err := Retry(ctx, func(ctx context.Context) error {
		var err error
		resp, err = client.DescribeImages(req)
		return err
	})
	if err != nil {
		return Halt(state, err, "Failed to get source image info")
	}

	if *resp.Response.TotalCount > 0 {
		state.Put("source_image", resp.Response.ImageSet[0])
		Message(state, *resp.Response.ImageSet[0].ImageName, "Image found")
		return multistep.ActionContinue
	}

	return Halt(state, fmt.Errorf("No image found under current instance_type(%s) restriction", config.InstanceType), "")
}

func (s *stepCheckSourceImage) Cleanup(bag multistep.StateBag) {}
