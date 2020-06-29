package lxd

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepPublish struct{
	client lxdClient
}

func (s *stepPublish) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	name := config.ContainerName

	ui.Say("Stopping container...")
	if err := s.client.StopContainer(name); err != nil {
		err := fmt.Errorf("Error stopping container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Publishing container...")
	fingerprint, err := s.client.PublishContainer(name, config.OutputImage, config.PublishProperties)
	if err != nil {
		err := fmt.Errorf("Error publishing container: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created image: %s", fingerprint))
	state.Put("imageFingerprint", fingerprint)

	return multistep.ActionContinue
}

func (s *stepPublish) Cleanup(state multistep.StateBag) {}
