package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepPreValidate struct {
	AlicloudDestImageName string
	ForceDelete           bool
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if err := s.validateRegions(state); err != nil {
		return halt(state, err, "")
	}

	if err := s.validateDestImageName(state); err != nil {
		return halt(state, err, "")
	}

	return multistep.ActionContinue
}

func (s *stepPreValidate) validateRegions(state multistep.StateBag) error {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	if config.AlicloudSkipValidation {
		ui.Say("Skip region validation flag found, skipping prevalidating source region and copied regions.")
		return nil
	}

	ui.Say("Prevalidating source region and copied regions...")

	var errs *packer.MultiError
	if err := config.ValidateRegion(config.AlicloudRegion); err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}
	for _, region := range config.AlicloudImageDestinationRegions {
		if err := config.ValidateRegion(region); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *stepPreValidate) validateDestImageName(state multistep.StateBag) error {
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)

	if s.ForceDelete {
		ui.Say("Force delete flag found, skipping prevalidating image name.")
		return nil
	}

	ui.Say("Prevalidating image name...")

	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.RegionId = config.AlicloudRegion
	describeImagesRequest.ImageName = s.AlicloudDestImageName
	describeImagesRequest.Status = ImageStatusQueried

	imagesResponse, err := client.DescribeImages(describeImagesRequest)
	if err != nil {
		return fmt.Errorf("Error querying alicloud image: %s", err)
	}

	images := imagesResponse.Images.Image
	if len(images) > 0 {
		return fmt.Errorf("Error: Image Name: '%s' is used by an existing alicloud image: %s", images[0].ImageName, images[0].ImageId)
	}

	return nil
}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {}
