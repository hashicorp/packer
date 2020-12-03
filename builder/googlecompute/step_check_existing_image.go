package googlecompute

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepCheckExistingImage represents a Packer build step that checks if the
// target image already exists, and aborts immediately if so.
type StepCheckExistingImage int

// Run executes the Packer build step that checks if the image already exists.
func (s *StepCheckExistingImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	d := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Checking image does not exist...")
	c.imageAlreadyExists = d.ImageExists(c.ImageName)
	if !c.PackerForce && c.imageAlreadyExists {
		err := fmt.Errorf("Image %s already exists.\n"+
			"Use the force flag to delete it prior to building.", c.ImageName)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepCheckExistingImage) Cleanup(state multistep.StateBag) {}
