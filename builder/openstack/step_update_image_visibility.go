package openstack

import (
	"context"
	"fmt"

	imageservice "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepUpdateImageVisibility struct{}

func (s *stepUpdateImageVisibility) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	imageId := state.Get("image").(string)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)

	if config.ImageVisibility == "" {
		return multistep.ActionContinue
	}
	imageClient, err := config.imageV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing image service client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Updating image visibility to %s", config.ImageVisibility))
	r := imageservice.Update(
		imageClient,
		imageId,
		imageservice.UpdateOpts{
			imageservice.UpdateVisibility{
				Visibility: config.ImageVisibility,
			},
		},
	)

	if _, err = r.Extract(); err != nil {
		err = fmt.Errorf("Error updating image visibility: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUpdateImageVisibility) Cleanup(multistep.StateBag) {
	// No cleanup...
}
