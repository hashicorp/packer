package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
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
//   full_disk_path (string) - The full path to the created disk.
type stepCreateDisk struct{}

func (stepCreateDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating virtual machine disk")
	full_disk_path := filepath.Join(config.OutputDir, config.DiskName+".vmdk")
	if err := driver.CreateDisk(full_disk_path, fmt.Sprintf("%dM", config.DiskSize), config.DiskTypeId); err != nil {
		err := fmt.Errorf("Error creating disk: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("full_disk_path", full_disk_path)

	return multistep.ActionContinue
}

func (stepCreateDisk) Cleanup(multistep.StateBag) {}
