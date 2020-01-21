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

func (s *stepPreMountCommands) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	device := state.Get("device").(string)

	ictx := config.ctx
	ictx.Data = &preMountCommandsData{
		Device:    device,
		MountPath: config.ChrootMountPath,
	}

	ui.Say("Running pre-mount commands...")
	if err := runCommands(config.PreMountCommands, ictx, state); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepPreMountCommands) Cleanup(state multistep.StateBag) {}
