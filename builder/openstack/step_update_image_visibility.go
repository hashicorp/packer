package openstack

import (
	"fmt"

	imageservice "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepUpdateImageVisibility struct{}

func (s *stepUpdateImageVisibility) Run(state multistep.StateBag) multistep.StepAction {
	imageId := state.Get("image").(string)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)

	if config.ImageVisibility != "" {
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

	}

	return multistep.ActionContinue
}

func (s *stepUpdateImageVisibility) Cleanup(multistep.StateBag) {
	// No cleanup...
}
