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
	SourceProperties map[string]string
}

func PropertiesSatisfied(image *images.Image, props *map[string]string) bool {
	for key, value := range *props {
		if image.Properties[key] != value {
			return false
		}
	}

	return true
}

func (s *StepSourceImageInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if s.SourceImage != "" {
		state.Put("source_image", s.SourceImage)

		return multistep.ActionContinue
	}

	client, err := config.imageV2Client()

	if s.SourceImageName != "" {
		s.SourceImageOpts = images.ListOpts{
			Name: s.SourceImageName,
		}
	}

	log.Printf("Using Image Filters %+v", s.SourceImageOpts)
	image := &images.Image{}
	err = images.List(client, s.SourceImageOpts).EachPage(func(page pagination.Page) (bool, error) {
		imgs, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}
		ui.Message(fmt.Sprintf("Resulting images: %d", len(imgs)))

		count := 0
		first := -1

		for index, img := range imgs {
			ui.Message(fmt.Sprintf("index +%v, image %+v", index, img))
			ui.Message(fmt.Sprintf("Metadata %+v", img.Metadata))
			ui.Message(fmt.Sprintf("Properties %+v", img.Properties))

			// Check if all Properties are satisfied
			if PropertiesSatisfied(&img, &s.SourceProperties) {
				ui.Message(fmt.Sprintf("Matched properties %+v", s.SourceProperties))
				count++
				if first < 0 {
					first = index
				}
				// Don't iterate over entries we will never use.
				if count > 1 {
					break
				}
			} else {
				ui.Message(fmt.Sprintf("FAILED to match properties %+v", s.SourceProperties))
			}
		}

		switch count {
		case 0:
			return true, nil
		case 1:
			*image = imgs[first]
			return false, nil
		default:
			if s.SourceMostRecent {
				*image = imgs[first]
				return false, nil
			}
			return false, fmt.Errorf(
				"Your query returned more than one result. Please try a more specific search, or set most_recent to true. Search filters: %+v",
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
		err := fmt.Errorf("No image was found matching filters: %+v", s.SourceImageOpts)
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
