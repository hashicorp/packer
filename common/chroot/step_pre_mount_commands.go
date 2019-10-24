package chroot

import (
	"context"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type preMountCommandsData struct {
	Device string
}

// StepPreMountCommands sets up the a new block device when building from scratch
type StepPreMountCommands struct {
	Commands []string
}

func (s *StepPreMountCommands) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(interpolateContextProvider)
	device := state.Get("device").(string)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	if len(s.Commands) == 0 {
		return multistep.ActionContinue
	}

	ictx := config.GetContext()
	ictx.Data = &preMountCommandsData{Device: device}

	ui.Say("Running device setup commands...")
	if err := RunLocalCommands(s.Commands, wrappedCommand, ictx, ui); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *StepPreMountCommands) Cleanup(state multistep.StateBag) {}
