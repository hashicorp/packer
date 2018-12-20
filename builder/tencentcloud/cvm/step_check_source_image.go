package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type stepCheckSourceImage struct {
	sourceImageId string
}

func (s *stepCheckSourceImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("cvm_client").(*cvm.Client)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	req := cvm.NewDescribeImagesRequest()
	req.ImageIds = []*string{&config.SourceImageId}
	req.InstanceType = &config.InstanceType

	resp, err := client.DescribeImages(req)
	if err != nil {
		err := fmt.Errorf("querying image info failed: %s", err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if *resp.Response.TotalCount > 0 { // public image or private image.
		state.Put("source_image", resp.Response.ImageSet[0])
		ui.Message(fmt.Sprintf("Image found: %s", *resp.Response.ImageSet[0].ImageId))
		return multistep.ActionContinue
	}
	// later market image will be included.
	err = fmt.Errorf("no image founded under current instance_type(%s) restriction", config.InstanceType)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func (s *stepCheckSourceImage) Cleanup(bag multistep.StateBag) {}
