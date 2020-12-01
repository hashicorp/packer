package chroot

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepCheckRootDevice makes sure the root device on the OMI is BSU-backed.
type StepCheckRootDevice struct{}

func (s *StepCheckRootDevice) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	image := state.Get("source_image").(osc.Image)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Checking the root device on source OMI...")

	// It must be BSU-backed otherwise the build won't work
	if image.RootDeviceType != "ebs" {
		err := fmt.Errorf("The root device of the source OMI must be BSU-backed.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCheckRootDevice) Cleanup(multistep.StateBag) {}
