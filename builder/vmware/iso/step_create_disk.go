package iso

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step creates the virtual disks for the VM.
//
// Uses:
//   config *config
//   driver Driver
//   ui     packer.Ui
//
// Produces:
//   disk_full_paths ([]string) - The full paths to all created disks
type stepCreateDisk struct{}

func (stepCreateDisk) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating required virtual machine disks")

	// Users can configure disks at several locations in the template so
	// first collate all the disk requirements
	var diskPaths, diskSizes []string
	// The 'main' or 'default' disk
	diskPaths = append(diskPaths, filepath.Join(config.OutputDir, config.DiskName+".vmdk"))
	diskSizes = append(diskSizes, fmt.Sprintf("%dM", uint64(config.DiskSize)))
	// Additional disks
	if len(config.AdditionalDiskSize) > 0 {
		for i, diskSize := range config.AdditionalDiskSize {
			path := filepath.Join(config.OutputDir, fmt.Sprintf("%s-%d.vmdk", config.DiskName, i+1))
			diskPaths = append(diskPaths, path)
			size := fmt.Sprintf("%dM", uint64(diskSize))
			diskSizes = append(diskSizes, size)
		}
	}

	// Create all required disks
	for i, diskPath := range diskPaths {
		log.Printf("[INFO] Creating disk with Path: %s and Size: %s", diskPath, diskSizes[i])
		// Additional disks currently use the same adapter type and disk
		// type as specified for the main disk
		if err := driver.CreateDisk(diskPath, diskSizes[i], config.DiskAdapterType, config.DiskTypeId); err != nil {
			err := fmt.Errorf("Error creating disk: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Stash the disk paths so we can retrieve later e.g. when compacting
	state.Put("disk_full_paths", diskPaths)
	return multistep.ActionContinue
}

func (stepCreateDisk) Cleanup(multistep.StateBag) {}
