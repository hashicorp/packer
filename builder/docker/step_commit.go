package docker

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCommit commits the container to a image.
type StepCommit struct {
	imageId string
}

func (s *StepCommit) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	containerId := state.Get("container_id").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Committing the container")
	imageId, err := driver.Commit(containerId)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Save the container ID
	s.imageId = imageId
	state.Put("image_id", s.imageId)
	ui.Message(fmt.Sprintf("Image ID: %s", s.imageId))

	return multistep.ActionContinue
}

func (s *StepCommit) Cleanup(state multistep.StateBag) {}
