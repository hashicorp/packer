package chroot

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepCheckRootDevice makes sure the root device on the OMI is BSU-backed.
type StepCheckRootDevice struct{}

func (s *StepCheckRootDevice) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	image := state.Get("source_image").(oapi.Image)
	ui := state.Get("ui").(packer.Ui)

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
