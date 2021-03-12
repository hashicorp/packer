package wim

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
)

type StepUnmountISO struct {
	DevicePathKey string
	ISOPathKey    string
	SkipOperation bool
}

func (s *StepUnmountISO) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.SkipOperation {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)
	isoPath := state.Get(s.ISOPathKey).(string)

	ui.Say("Unmounting ISO...")

	if err := s.dismountDiskImage(isoPath); err != nil {
		err = fmt.Errorf("Error dismounting ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Update state bag
	state.Remove(s.DevicePathKey)

	return multistep.ActionContinue
}

// Cleanup does nothing
func (s *StepUnmountISO) Cleanup(state multistep.StateBag) {}

// TODO: dedup
func (s *StepUnmountISO) dismountDiskImage(isoPath string) error {

	var script = `
param([string]$imagePath)
Dismount-DiskImage -ImagePath $imagePath
`

	var ps powershell.PowerShellCmd
	return ps.Run(script, isoPath)
}
