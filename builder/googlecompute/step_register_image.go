package googlecompute

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepRegisterImage represents a Packer build step that registers GCE machine images.
type StepRegisterImage int

// Run executes the Packer build step that registers a GCE machine image.
func (s *StepRegisterImage) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	var err error
	imageURL := fmt.Sprintf(
		"https://storage.cloud.google.com/%s/%s.tar.gz",
		config.BucketName, config.ImageName)

	ui.Say("Registering image...")
	errCh := driver.CreateImage(config.ImageName, config.ImageDescription, imageURL)
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

	state.Put("image_name", config.ImageName)
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepRegisterImage) Cleanup(state multistep.StateBag) {}
