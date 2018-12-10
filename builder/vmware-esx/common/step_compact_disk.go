package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step compacts the virtual disk for the VM unless the "skip_compaction"
// boolean is true.
//
// Uses:
//   driver Driver
//   disk_full_paths ([]string) - The full paths to all created disks
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type StepCompactDisk struct {
	Skip bool
}

func (s StepCompactDisk) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	diskFullPaths := state.Get("disk_full_paths").([]string)

	if s.Skip {
		log.Println("Skipping disk compaction step...")
		return multistep.ActionContinue
	}

	ui.Say("Compacting all attached virtual disks...")
	for i, diskFullPath := range diskFullPaths {
		ui.Message(fmt.Sprintf("Compacting virtual disk %d", i+1))
		if err := driver.CompactDisk(diskFullPath); err != nil {
			state.Put("error", fmt.Errorf("Error compacting disk: %s", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (StepCompactDisk) Cleanup(multistep.StateBag) {}
