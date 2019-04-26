package docker

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepCommit commits the container to a image.
type StepCommit struct {
	imageId string
}

func (s *StepCommit) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	containerId := state.Get("container_id").(string)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if config.WindowsContainer {
		// docker can't commit a running Windows container
		err := driver.StopContainer(containerId)
		if err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error halting windows container for commit: %s",
				err.Error()))
			return multistep.ActionHalt
		}
	}
	ui.Say("Committing the container")
	imageId, err := driver.Commit(containerId, config.Author, config.Changes, config.Message)
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
