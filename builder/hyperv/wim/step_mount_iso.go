package wim

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
)

type StepMountISO struct {
	DevicePathKey string
	ISOPathKey    string
	SkipOperation bool
}

func (s *StepMountISO) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.SkipOperation {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)
	isoPath := state.Get(s.ISOPathKey).(string)

	ui.Say("Mounting ISO...")

	// Mount ISO
	devicePath, err := s.mountDiskImage(isoPath)
	if err != nil {
		err = fmt.Errorf("Error mounting ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Update state bag
	state.Put(s.DevicePathKey, devicePath)

	return multistep.ActionContinue
}

func (s *StepMountISO) Cleanup(state multistep.StateBag) {
	if s.SkipOperation {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	// Dismount ISO
	if _, ok := state.GetOk(s.DevicePathKey); ok {
		isoPath := state.Get(s.ISOPathKey).(string)

		if err := s.dismountDiskImage(isoPath); err != nil {
			err = fmt.Errorf("Error dismounting ISO: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
		} else {
			// Update state bag
			state.Remove(s.DevicePathKey)
		}
	}
}

func (s *StepMountISO) dismountDiskImage(isoPath string) error {

	var script = `
param([string]$imagePath)
Dismount-DiskImage -ImagePath $imagePath
`

	var ps powershell.PowerShellCmd
	return ps.Run(script, isoPath)
}

func (s *StepMountISO) mountDiskImage(isoPath string) (string, error) {

	var script = `
param([string]$imagePath)
$diskImage = Mount-DiskImage -ImagePath $imagePath
if ($diskImage -ne $null) {
    $diskImage.DevicePath
}
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, isoPath)

	if err != nil {
		return "", err
	}

	var devicePath = strings.TrimSpace(cmdOut)
	return devicePath, err
}
