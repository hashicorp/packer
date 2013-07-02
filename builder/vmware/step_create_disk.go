package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
)

// This step creates the virtual disks for the VM.
//
// Uses:
//   config *config
//   driver Driver
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepCreateDisk struct{}

func (stepCreateDisk) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	ui.Say("Creating virtual machine disk")
	full_disk_path := filepath.Join(config.OutputDir, config.DiskName+".vmdk")
	state["full_disk_path"] = full_disk_path
	if err := driver.CreateDisk(full_disk_path, fmt.Sprintf("%dM", config.DiskSize)); err != nil {
		err := fmt.Errorf("Error creating disk: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCreateDisk) Cleanup(map[string]interface{}) {}
