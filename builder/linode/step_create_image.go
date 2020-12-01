package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
)

type stepCreateImage struct {
	client linodego.Client
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	disk := state.Get("disk").(*linodego.InstanceDisk)
	instance := state.Get("instance").(*linodego.Instance)

	ui.Say("Creating image...")
	image, err := s.client.CreateImage(ctx, linodego.ImageCreateOptions{
		DiskID:      disk.ID,
		Label:       c.ImageLabel,
		Description: c.Description,
	})

	if err == nil {
		_, err = s.client.WaitForInstanceDiskStatus(ctx, instance.ID, disk.ID, linodego.DiskReady, 600)
	}

	if err == nil {
		image, err = s.client.GetImage(ctx, image.ID)
	}

	if err != nil {
		err = errors.New("Error creating image: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("image", image)
	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}
