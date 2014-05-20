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

type stepAddDisks struct{}

func (stepAddDisks) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	full_disk_paths := state.Get("full_disk_paths").([]string)

	ui.Say("Creating additional virtual machine disks...")
	for i, args := range config.AddDisksConfig.NewDisks {
		full_disk_path := filepath.Join(config.OutputDir, args[0]+".vmdk")
		if err := driver.CreateDisk(full_disk_path, fmt.Sprintf("%sM", args[1]), args[2]); err != nil {
			err := fmt.Errorf("Error creating disk[%d]: %s", i, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		full_disk_paths = append(full_disk_paths, full_disk_path)

	}

	state.Put("full_disk_paths", full_disk_paths)
	return multistep.ActionContinue
}

func (stepAddDisks) Cleanup(multistep.StateBag) {}
