package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step compacts the virtual disk for the VM.
//
// Uses:
//   driver Driver
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepCompactDisk struct{}

func (stepCompactDisk) Run(state map[string]interface{}) multistep.StepAction {
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	full_disk_path := state["full_disk_path"].(string)

	ui.Say("Compacting the disk image")
	if err := driver.CompactDisk(full_disk_path); err != nil {
		state["error"] := fmt.Errorf("Error compacting disk: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCompactDisk) Cleanup(map[string]interface{}) {}
