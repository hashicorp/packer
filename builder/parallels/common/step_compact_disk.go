package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step removes all empty blocks from expanding Parallels virtual disks
// and reduces the result disk size
//
// Uses:
//   driver Driver
//   vmName string
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type StepCompactDisk struct {
	Skip bool
}

func (s *StepCompactDisk) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packer.Ui)

	if s.Skip {
		ui.Say("Skipping disk compaction step...")
		return multistep.ActionContinue
	}

	ui.Say("Compacting the disk image")
	diskPath, err := driver.DiskPath(vmName)
	if err != nil {
		err := fmt.Errorf("Error detecting virtual disk path: %s", err)
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

func (*StepCompactDisk) Cleanup(multistep.StateBag) {}
