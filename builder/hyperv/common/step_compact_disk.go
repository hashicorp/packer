package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCompactDisk struct {
	SkipCompaction bool
}

// Run runs a compaction/optimisation process on attached VHD/VHDX disks
func (s *StepCompactDisk) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if s.SkipCompaction {
		ui.Say("Skipping disk compaction...")
		return multistep.ActionContinue
	}

	// Get the tmp dir used to store the VMs files during the build process
	tmpPath := state.Get("packerTempDir").(string)

	ui.Say("Compacting disks...")
	// CompactDisks searches for all VHD/VHDX files under the supplied
	// path and runs the compacting process on each of them
	err := driver.CompactDisks(tmpPath)
	if err != nil {
		err := fmt.Errorf("Error compacting disks: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup does nothing
func (s *StepCompactDisk) Cleanup(state multistep.StateBag) {}
