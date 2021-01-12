package cvm

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCheckSourceImage struct {
	sourceImageId string
}

func (s *stepCheckSourceImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		imageNameRegex *regexp.Regexp
		err            error
	)
	config := state.Get("config").(*Config)
	client := state.Get("cvm_client").(*cvm.Client)

	Say(state, config.SourceImageId, "Trying to check source image")

	req := cvm.NewDescribeImagesRequest()
	req.InstanceType = &config.InstanceType
	if config.SourceImageId != "" {
		req.ImageIds = []*string{&config.SourceImageId}
	} else {
		imageNameRegex, err = regexp.Compile(config.SourceImageName)
		if err != nil {
			return Halt(state, fmt.Errorf("regex compilation error"), "Bad input")
		}
	}
	var resp *cvm.DescribeImagesResponse
	err = Retry(ctx, func(ctx context.Context) error {
		var err error
		resp, err = client.DescribeImages(req)
		return err
	})
	if err != nil {
		return Halt(state, err, "Failed to get source image info")
	}

	if *resp.Response.TotalCount > 0 {
		images := resp.Response.ImageSet
		if imageNameRegex != nil {
			for _, image := range images {
				if imageNameRegex.MatchString(*image.ImageName) {
					state.Put("source_image", image)
					Message(state, *image.ImageName, "Image found")
					return multistep.ActionContinue
				}
			}
		} else {
			state.Put("source_image", images[0])
			Message(state, *resp.Response.ImageSet[0].ImageName, "Image found")
			return multistep.ActionContinue
		}
	}

	return Halt(state, fmt.Errorf("No image found under current instance_type(%s) restriction", config.InstanceType), "")
}

func (s *stepCheckSourceImage) Cleanup(bag multistep.StateBag) {}
