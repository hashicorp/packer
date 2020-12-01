package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/members"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepAddImageMembers struct{}

func (s *stepAddImageMembers) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	imageId := state.Get("image").(string)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	if len(config.ImageMembers) == 0 {
		return multistep.ActionContinue
	}

	imageClient, err := config.imageV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing image service client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	for _, member := range config.ImageMembers {
		ui.Say(fmt.Sprintf("Adding member '%s' to image %s", member, imageId))
		r := members.Create(imageClient, imageId, member)
		if _, err = r.Extract(); err != nil {
			err = fmt.Errorf("Error adding member to image: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	if config.ImageAutoAcceptMembers {
		for _, member := range config.ImageMembers {
			ui.Say(fmt.Sprintf("Accepting image %s for member '%s'", imageId, member))
			r := members.Update(imageClient, imageId, member, members.UpdateOpts{Status: "accepted"})
			if _, err = r.Extract(); err != nil {
				err = fmt.Errorf("Error accepting image for member: %s", err)
				state.Put("error", err)
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepAddImageMembers) Cleanup(multistep.StateBag) {
	// No cleanup...
}
