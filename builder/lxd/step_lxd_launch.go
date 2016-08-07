package lxd

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

type stepLxdLaunch struct{}

func (s *stepLxdLaunch) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName
	image := config.Image

	args := []string{
		"launch", "--ephemeral=false", image, name,
	}

	ui.Say("Creating container...")
	_, err := LXDCommand(args...)
	if err != nil {
		err := fmt.Errorf("Error creating container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// TODO: Should we check `lxc info <container>` for "Running"?
	// We have to do this so /tmp doens't get cleared and lose our provisioner scripts.
	time.Sleep(1 * time.Second)

	return multistep.ActionContinue
}

func (s *stepLxdLaunch) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	args := []string{
		"delete", "--force", config.ContainerName,
	}

	ui.Say("Unregistering and deleting deleting container...")
	if _, err := LXDCommand(args...); err != nil {
		ui.Error(fmt.Sprintf("Error deleting container: %s", err))
	}
}
