package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepCompactDisk is a step that removes all empty blocks from expanding
// Parallels virtual disks and reduces the result disk size
//
// Uses:
//   driver Driver
//   vmName string
//   ui     packersdk.Ui
//
// Produces:
//   <nothing>
type StepCompactDisk struct {
	Skip bool
}

// Run runs the compaction of the virtual disk attached to the VM.
func (s *StepCompactDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packersdk.Ui)

	if s.Skip {
		ui.Say("Skipping disk compaction step...")
		return multistep.ActionContinue
	}

	ui.Say("Compacting the disk image")
	diskPath, err := driver.DiskPath(vmName)
	if err != nil {
		err = fmt.Errorf("Error detecting virtual disk path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := driver.CompactDisk(diskPath); err != nil {
		state.Put("error", fmt.Errorf("Error compacting disk: %s", err))
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup does nothing.
func (*StepCompactDisk) Cleanup(multistep.StateBag) {}
