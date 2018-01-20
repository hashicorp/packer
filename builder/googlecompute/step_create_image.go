package googlecompute

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepCreateImage represents a Packer build step that creates GCE machine
// images.
type StepCreateImage int

// Run executes the Packer build step that creates a GCE machine image.
//
// The image is created from the persistent disk used by the instance. The
// instance must be deleted and the disk retained before doing this step.
func (s *StepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if config.PackerForce && config.imageAlreadyExists {
		ui.Say("Deleting previous image...")

		errCh := driver.DeleteImage(config.ImageName)
		err := <-errCh
		if err != nil {
			err := fmt.Errorf("Error deleting image: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Creating image...")

	imageCh, errCh := driver.CreateImage(
		config.ImageName, config.ImageDescription, config.ImageFamily, config.Zone,
		config.DiskName, config.ImageLabels)
	var err error
	select {
	case err = <-errCh:
	case <-time.After(config.stateTimeout):
		err = errors.New("time out while waiting for image to register")
	}

	if err != nil {
		err := fmt.Errorf("Error waiting for image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("image", <-imageCh)
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepCreateImage) Cleanup(state multistep.StateBag) {}
