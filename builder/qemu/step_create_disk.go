package qemu

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	name := config.VMName

	if config.DiskImage && !config.UseBackingFile {
		return multistep.ActionContinue
	}

	var diskFullPaths, diskSizes []string

	ui.Say("Creating required virtual machine disks")
	// The 'main' or 'default' disk
	diskFullPaths = append(diskFullPaths, filepath.Join(config.OutputDir, name))
	diskSizes = append(diskSizes, config.DiskSize)
	// Additional disks
	if len(config.AdditionalDiskSize) > 0 {
		for i, diskSize := range config.AdditionalDiskSize {
			path := filepath.Join(config.OutputDir, fmt.Sprintf("%s-%d", name, i+1))
			diskFullPaths = append(diskFullPaths, path)
			size := diskSize
			diskSizes = append(diskSizes, size)
		}
	}

	// Create all required disks
	for i, diskFullPath := range diskFullPaths {
		log.Printf("[INFO] Creating disk with Path: %s and Size: %s", diskFullPath, diskSizes[i])
		command := []string{
			"create",
			"-f", config.Format,
		}

		if config.UseBackingFile && i == 0 {
			isoPath := state.Get("iso_path").(string)
			command = append(command, "-b", isoPath)
		}

		command = append(command,
			diskFullPath,
			diskSizes[i])

		if err := driver.QemuImg(command...); err != nil {
			err := fmt.Errorf("Error creating hard drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Stash the disk paths so we can retrieve later
	state.Put("qemu_disk_paths", diskFullPaths)

	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
