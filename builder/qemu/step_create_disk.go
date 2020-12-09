package qemu

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct {
	AdditionalDiskSize []string
	DiskImage          bool
	DiskSize           string
	Format             string
	OutputDir          string
	UseBackingFile     bool
	VMName             string
	QemuImgArgs        QemuImgArgs
}

func (s *stepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	name := s.VMName

	if len(s.AdditionalDiskSize) > 0 || s.UseBackingFile {
		ui.Say("Creating required virtual machine disks")
	}

	// The 'main' or 'default' disk
	diskFullPaths := []string{filepath.Join(s.OutputDir, name)}
	diskSizes := []string{s.DiskSize}

	// Additional disks
	if len(s.AdditionalDiskSize) > 0 {
		for i, diskSize := range s.AdditionalDiskSize {
			path := filepath.Join(s.OutputDir, fmt.Sprintf("%s-%d", name, i+1))
			diskFullPaths = append(diskFullPaths, path)
			diskSizes = append(diskSizes, diskSize)
		}
	}

	// Create all required disks
	for i, diskFullPath := range diskFullPaths {
		if s.DiskImage && !s.UseBackingFile && i == 0 {
			// Let the copy disk step (step_copy_disk.go) create the 'main' or
			// 'default' disk.
			continue
		}
		log.Printf("[INFO] Creating disk with Path: %s and Size: %s", diskFullPath, diskSizes[i])

		command := s.buildCreateCommand(diskFullPath, diskSizes[i], i, state)

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

func (s *stepCreateDisk) buildCreateCommand(path string, size string, i int, state multistep.StateBag) []string {
	command := []string{"create", "-f", s.Format}

	if s.DiskImage && s.UseBackingFile && i == 0 {
		// Use a backing file for the 'main' or 'default' disk
		isoPath := state.Get("iso_path").(string)
		command = append(command, "-b", isoPath)
	}

	// add user-provided convert args
	command = append(command, s.QemuImgArgs.Create...)

	// add target path and size.
	command = append(command, path, size)

	return command
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
