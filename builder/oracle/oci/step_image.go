package oci

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepImage struct{}

func (s *stepImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		driver     = state.Get("driver").(Driver)
		ui         = state.Get("ui").(packersdk.Ui)
		instanceID = state.Get("instance_id").(string)
	)

	ui.Say("Creating image from instance...")

	image, err := driver.CreateImage(ctx, instanceID)
	if err != nil {
		err = fmt.Errorf("Error creating image from instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	err = driver.WaitForImageCreation(ctx, *image.Id)
	if err != nil {
		err = fmt.Errorf("Error waiting for image creation to finish: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// TODO(apryde): This is stale as .LifecycleState has changed to
	// AVAILABLE at this point. Does it matter?
	state.Put("image", image)

	ui.Say("Image created.")

	return multistep.ActionContinue
}

func (s *stepImage) Cleanup(state multistep.StateBag) {
	// Nothing to do
}
