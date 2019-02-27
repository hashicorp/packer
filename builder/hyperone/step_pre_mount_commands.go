package hyperone

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type preMountCommandsData struct {
	Device    string
	MountPath string
}

type stepPreMountCommands struct{}

func (s *stepPreMountCommands) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	device := state.Get("device").(string)

	ctx := config.ctx
	ctx.Data = &preMountCommandsData{
		Device:    device,
		MountPath: config.ChrootMountPath,
	}

	ui.Say("Running pre-mount commands...")
	if err := runCommands(config.PreMountCommands, ctx, state); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepPreMountCommands) Cleanup(state multistep.StateBag) {}
