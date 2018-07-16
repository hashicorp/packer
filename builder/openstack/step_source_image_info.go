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
	SourceImage     string
	SourceImageName string
	ImageFilters    ImageFilterOptions
}

type ImageFilterOptions struct {
	Filters    map[string]string `mapstructure:"filters"`
	MostRecent bool              `mapstructure:"most_recent"`
}

func (s *StepSourceImageInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	client, err := config.computeV2Client()

	// if an ID is provided we skip the filter since that will return a single image
	if s.SourceImage != "" {
		state.Put("source_image", s.SourceImage)
		return multistep.ActionContinue
	}

	params := &images.ListOpts{}

	// build ListOpts from filters
	if len(s.ImageFilters.Filters) > 0 {
		err = buildImageFilters(s.ImageFilters.Filters, params)
		if err != nil {
			err := fmt.Errorf("Errors encountered in filter parsing.\n%s" + err.Error())
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if s.ImageFilters.MostRecent {
		applyMostRecent(params)
	}

	log.Printf("Using Image Filters %v", params)
	image := &images.Image{}
	err = images.List(client, params).EachPage(func(page pagination.Page) (bool, error) {
		i, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		switch len(i) {
		case 0:
			return false, fmt.Errorf("No image was found matching filters: %v", params)
		case 1:
			*image = i[0]
			return true, nil
		default:
			return false, fmt.Errorf(
				"Your query returned more than one result. Please try a more specific search, or set most_recent to true. Search filters: %v",
				params)
		}
	})

	if err != nil {
		err := fmt.Errorf("Error querying image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Found Image ID: %s", image.ID))

	state.Put("source_image", image)
	return multistep.ActionContinue
}

func (s *StepSourceImageInfo) Cleanup(state multistep.StateBag) {
	// No cleanup required for backout
}
