package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
)

type stepShutdownLinode struct {
	client linodego.Client
}

func (s *stepShutdownLinode) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	instance := state.Get("instance").(*linodego.Instance)

	ui.Say("Shutting down Linode...")
	if err := s.client.ShutdownInstance(ctx, instance.ID); err != nil {
		err = errors.New("Error shutting down Linode: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	_, err := s.client.WaitForInstanceStatus(ctx, instance.ID, linodego.InstanceOffline, int(c.StateTimeout.Seconds()))
	if err != nil {
		err = errors.New("Error shutting down Linode: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepShutdownLinode) Cleanup(state multistep.StateBag) {}
