package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
)

type stepCheckAccountImageLimit struct {
	client linodego.Client
}

func (s *stepCheckAccountImageLimit) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	accountImageLimit := c.AccountImageLimit

	if accountImageLimit > 0 {
		ui.Say("Checking image count...")
		images, err := s.client.ListImages(ctx, nil)
		if err != nil {
			err = errors.New("Error listing Images: " + err.Error())
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(images) >= accountImageLimit {
			err = errors.New("Account Image Limit reached.")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepCheckAccountImageLimit) Cleanup(state multistep.StateBag) {}
