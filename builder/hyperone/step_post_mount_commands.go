package hyperone

import (
	"context"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type postMountCommandsData struct {
	Device    string
	MountPath string
}

type stepPostMountCommands struct{}

func (s *stepPostMountCommands) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	device := state.Get("device").(string)

	ictx := config.ctx
	ictx.Data = &postMountCommandsData{
		Device:    device,
		MountPath: config.ChrootMountPath,
	}

	ui.Say("Running post-mount commands...")
	if err := runCommands(config.PostMountCommands, ictx, state); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepPostMountCommands) Cleanup(state multistep.StateBag) {}
