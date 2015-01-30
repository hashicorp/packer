package googlecompute

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCheckExistingImage represents a Packer build step that checks if the
// target image already exists, and aborts immediately if so.
type StepCheckExistingImage int

// Run executes the Packer build step that checks if the image already exists.
func (s *StepCheckExistingImage) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Checking image does not exist...")
	exists := driver.ImageExists(config.ImageName)
	if exists {
		err := fmt.Errorf("Image %s already exists", config.ImageName)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup.
func (s *StepCheckExistingImage) Cleanup(state multistep.StateBag) {}
