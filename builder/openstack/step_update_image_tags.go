package openstack

import (
	"context"
	"fmt"
	"strings"

	imageservice "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepUpdateImageTags struct{}

func (s *stepUpdateImageTags) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	imageId := state.Get("image").(string)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	if len(config.ImageTags) == 0 {
		return multistep.ActionContinue
	}
	imageClient, err := config.imageV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing image service client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Updating image tags to %s", strings.Join(config.ImageTags, ", ")))
	r := imageservice.Update(
		imageClient,
		imageId,
		imageservice.UpdateOpts{
			imageservice.ReplaceImageTags{
				NewTags: config.ImageTags,
			},
		},
	)

	if _, err = r.Extract(); err != nil {
		err = fmt.Errorf("Error updating image tags: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUpdateImageTags) Cleanup(multistep.StateBag) {
	// No cleanup...
}
