package qemu

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step copies the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCopyDisk struct {
	DiskImage      bool
	Format         string
	OutputDir      string
	UseBackingFile bool
	VMName         string

	QemuImgArgs QemuImgArgs
}

func (s *stepCopyDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packersdk.Ui)
	path := filepath.Join(s.OutputDir, s.VMName)

	// Only make a copy of the ISO as the 'main' or 'default' disk if is a disk
	// image and not using a backing file. The create disk step (step_create_disk.go)
	// would already have made the disk otherwise
	if !s.DiskImage || s.UseBackingFile {
		return multistep.ActionContinue
	}

	// In some cases, the file formats provided are equivalent by comparing the
	// file extensions. Skip the conversion step
	// This also serves as a workaround for a QEMU bug: https://bugs.launchpad.net/qemu/+bug/1776920
	ext := filepath.Ext(isoPath)
	if len(ext) >= 1 && ext[1:] == s.Format && len(s.QemuImgArgs.Convert) == 0 {
		ui.Message("File extension already matches desired output format. " +
			"Skipping qemu-img convert step")
		err := driver.Copy(isoPath, path)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		return multistep.ActionContinue
	}

	command := s.buildConvertCommand(isoPath, path)

	ui.Say("Copying hard drive...")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCopyDisk) buildConvertCommand(sourcePath, targetPath string) []string {
	command := []string{"convert"}

	// Add user-provided convert args
	command = append(command, s.QemuImgArgs.Convert...)

	// Add format, and paths.
	command = append(command, "-O", s.Format, sourcePath, targetPath)

	return command
}

func (s *stepCopyDisk) Cleanup(state multistep.StateBag) {}
