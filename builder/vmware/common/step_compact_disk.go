package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step compacts the virtual disk for the VM unless the "skip_compaction"
// boolean is true.
//
// Uses:
//   driver Driver
//   full_disk_path string
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type StepCompactDisk struct {
	Skip bool
}

func (s StepCompactDisk) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	full_disk_paths := state.Get("full_disk_paths").([]string)

	if s.Skip {
		log.Println("Skipping disk compaction step...")
		return multistep.ActionContinue
	}

	ui.Say("Compacting the disk image(s)")
	for i, disk := range full_disk_paths {
		if err := driver.CompactDisk(disk); err != nil {
			state.Put("error", fmt.Errorf("Error compacting disk[%d]: %s", i, err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (StepCompactDisk) Cleanup(multistep.StateBag) {}
