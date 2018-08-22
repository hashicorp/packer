package openstack

import (
	"context"
	"fmt"
	"log"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSourceImageInfo struct {
	SourceImage      string
	SourceImageName  string
	SourceImageOpts  images.ListOpts
	SourceMostRecent bool
}

func (s *StepSourceImageInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	if s.SourceImage != "" || s.SourceImageName != "" {
		return multistep.ActionContinue
	}

	client, err := config.imageV2Client()

	log.Printf("Using Image Filters %v", s.SourceImageOpts)
	image := &images.Image{}
	err = images.List(client, s.SourceImageOpts).EachPage(func(page pagination.Page) (bool, error) {
		i, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		switch len(i) {
		case 1:
			*image = i[0]
			return false, nil
		default:
			if s.SourceMostRecent {
				*image = i[0]
				return false, nil
			}
			return false, fmt.Errorf(
				"Your query returned more than one result. Please try a more specific search, or set most_recent to true. Search filters: %v",
				s.SourceImageOpts)
		}
	})

	if err != nil {
		err := fmt.Errorf("Error querying image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if image.ID == "" {
		err := fmt.Errorf("No image was found matching filters: %v", s.SourceImageOpts)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Found Image ID: %s", image.ID))

	state.Put("source_image", image.ID)
	return multistep.ActionContinue
}

func (s *StepSourceImageInfo) Cleanup(state multistep.StateBag) {
	// No cleanup required for backout
}
