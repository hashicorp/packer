package hyperone

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type postMountCommandsData struct {
	Device    string
	MountPath string
}

type stepPostMountCommands struct{}

func (s *stepPostMountCommands) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	device := state.Get("device").(string)

	ctx := config.ctx
	ctx.Data = &postMountCommandsData{
		Device:    device,
		MountPath: config.ChrootMountPath,
	}

	ui.Say("Running post-mount commands...")
	if err := runCommands(config.PostMountCommands, ctx, state); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepPostMountCommands) Cleanup(state multistep.StateBag) {}
