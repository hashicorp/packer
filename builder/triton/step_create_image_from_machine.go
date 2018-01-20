package triton

import (
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepCreateImageFromMachine creates an image with the specified attributes
// from the machine with the given ID, and waits for the image to be created.
// The machine must be in the "stopped" state prior to this step being run.
type StepCreateImageFromMachine struct{}

func (s *StepCreateImageFromMachine) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	machineId := state.Get("machine").(string)

	ui.Say("Creating image from source machine...")

	imageId, err := driver.CreateImageFromMachine(machineId, config)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem creating image from machine: %s", err))
		return multistep.ActionHalt
	}

	ui.Say("Waiting for image to become available...")
	err = driver.WaitForImageCreation(imageId, 10*time.Minute)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem waiting for image to become available: %s", err))
		return multistep.ActionHalt
	}

	state.Put("image", imageId)

	return multistep.ActionContinue
}

func (s *StepCreateImageFromMachine) Cleanup(state multistep.StateBag) {
	// No cleanup
}
