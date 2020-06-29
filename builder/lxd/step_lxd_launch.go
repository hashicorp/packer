package lxd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepLxdLaunch struct{
	client lxdClient
}

func (s *stepLxdLaunch) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating container...")
	if err := s.client.LaunchContainer(config.ContainerName, config.Image, config.Profile, config.LaunchConfig); err != nil {
		err := fmt.Errorf("Error creating container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	sleep_seconds, err := strconv.Atoi(config.InitSleep)
	if err != nil {
		err := fmt.Errorf("Error parsing InitSleep into int: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// TODO: Should we check `lxc info <container>` for "Running"?
	// We have to do this so /tmp doesn't get cleared and lose our provisioner scripts.

	time.Sleep(time.Duration(sleep_seconds) * time.Second)
	log.Printf("Sleeping for %d seconds...", sleep_seconds)
	return multistep.ActionContinue
}

func (s *stepLxdLaunch) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering and deleting container...")
	if err := s.client.DeleteContainer(config.ContainerName); err != nil {
		ui.Error(fmt.Sprintf("Error deleting container: %s", err))
	}
}

