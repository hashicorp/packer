package openstack

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imageimport"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepSourceImageInfo struct {
	SourceImage               string
	SourceImageName           string
	ExternalSourceImageURL    string
	ExternalSourceImageFormat string
	SourceImageOpts           images.ListOpts
	SourceMostRecent          bool
	SourceProperties          map[string]string
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
	ui := state.Get("ui").(packersdk.Ui)

	client, err := config.imageV2Client()
	if err != nil {
		err := fmt.Errorf("error creating image client: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if s.ExternalSourceImageURL != "" {
		createOpts := images.CreateOpts{
			Name:            s.SourceImageName,
			ContainerFormat: "bare",
			DiskFormat:      s.ExternalSourceImageFormat,
			Properties: map[string]string{
				"packer_external_source_image_url":    s.ExternalSourceImageURL,
				"packer_external_source_image_format": s.ExternalSourceImageFormat,
			},
		}

		ui.Say("Creating image using external source image with name " + s.SourceImageName)
		ui.Say("Using disk format " + s.ExternalSourceImageFormat)
		image, err := images.Create(client, createOpts).Extract()

		if err != nil {
			err := fmt.Errorf("Error creating source image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Say("Created image with ID " + image.ID)

		importOpts := imageimport.CreateOpts{
			Name: imageimport.WebDownloadMethod,
			URI:  s.ExternalSourceImageURL,
		}

		ui.Say("Importing External Source Image from URL " + s.ExternalSourceImageURL)
		err = imageimport.Create(client, image.ID, importOpts).ExtractErr()

		if err != nil {
			err := fmt.Errorf("Error importing source image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		for image.Status != images.ImageStatusActive {
			ui.Message("Image not Active, retrying in 10 seconds")
			time.Sleep(10 * time.Second)

			img, err := images.Get(client, image.ID).Extract()

			if err != nil {
				err := fmt.Errorf("Error querying image: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			image = img
		}

		s.SourceImage = image.ID
	}

	if s.SourceImage != "" {
		state.Put("source_image", s.SourceImage)

		return multistep.ActionContinue
	}

	if s.SourceImageName != "" {
		s.SourceImageOpts = images.ListOpts{
			Name: s.SourceImageName,
		}
	}

	log.Printf("Using Image Filters %+v", s.SourceImageOpts)
	image := &images.Image{}
	count := 0
	err = images.List(client, s.SourceImageOpts).EachPage(func(page pagination.Page) (bool, error) {
		imgs, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, img := range imgs {
			// Check if all Properties are satisfied
			if PropertiesSatisfied(&img, &s.SourceProperties) {
				count++
				if count == 1 {
					// Tentatively return this result.
					*image = img
				}
				// Don't iterate over entries we will never use.
				if count > 1 {
					break
				}
			}
		}

		switch count {
		case 0: // Continue looking at next page.
			return true, nil
		case 1: // Maybe we're done, maybe there is another result in a later page and it is an error.
			if s.SourceMostRecent {
				return false, nil
			}
			return true, nil
		default: // By now we should know if getting 2+ results is an error or not.
			if s.SourceMostRecent {
				return false, nil
			}
			return false, fmt.Errorf(
				"Your query returned more than one result. Please try a more specific search, or set most_recent to true. Search filters: %+v properties %+v",
				s.SourceImageOpts, s.SourceProperties)
		}
	})

	if err != nil {
		err := fmt.Errorf("Error querying image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if image.ID == "" {
		err := fmt.Errorf("No image was found matching filters: %+v properties %+v",
			s.SourceImageOpts, s.SourceProperties)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Found Image ID: %s", image.ID))

	state.Put("source_image", image.ID)
	return multistep.ActionContinue
}

func (s *StepSourceImageInfo) Cleanup(state multistep.StateBag) {
	if s.ExternalSourceImageURL != "" {
		config := state.Get("config").(*Config)
		ui := state.Get("ui").(packersdk.Ui)

		client, err := config.imageV2Client()
		if err != nil {
			err := fmt.Errorf("error creating image client: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return
		}

		ui.Say(fmt.Sprintf("Deleting temporary external source image: %s ...", s.SourceImageName))
		err = images.Delete(client, s.SourceImage).ExtractErr()
		if err != nil {
			err := fmt.Errorf("error cleaning up external source image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return
		}
	}
}
