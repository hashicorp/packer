package chroot

import (
	"context"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

type postMountCommandsData struct {
	Device    string
	MountPath string
}

// StepPostMountCommands allows running arbitrary commands after mounting the
// device, but prior to the bind mount and copy steps.
type StepPostMountCommands struct {
	Commands []string
}

func (s *StepPostMountCommands) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(interpolateContextProvider)
	device := state.Get("device").(string)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	if len(s.Commands) == 0 {
		return multistep.ActionContinue
	}

	ictx := config.GetContext()
	ictx.Data = &postMountCommandsData{
		Device:    device,
		MountPath: mountPath,
	}

	ui.Say("Running post-mount commands...")
	if err := RunLocalCommands(s.Commands, wrappedCommand, ictx, ui); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *StepPostMountCommands) Cleanup(state multistep.StateBag) {}
