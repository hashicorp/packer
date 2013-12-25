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
	full_disk_path := state.Get("full_disk_path").(string)

	if s.Skip {
		log.Println("Skipping disk compaction step...")
		return multistep.ActionContinue
	}

	ui.Say("Compacting the disk image")
	if err := driver.CompactDisk(full_disk_path); err != nil {
		state.Put("error", fmt.Errorf("Error compacting disk: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (StepCompactDisk) Cleanup(multistep.StateBag) {}
