package qemu

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step resizes the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepResizeDisk struct {
	DiskCompression bool
	DiskImage       bool
	Format          string
	OutputDir       string
	SkipResizeDisk  bool
	VMName          string
	DiskSize        string

	QemuImgArgs QemuImgArgs
}

func (s *stepResizeDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	path := filepath.Join(s.OutputDir, s.VMName)

	command := s.buildResizeCommand(path)

	if s.DiskImage == false || s.SkipResizeDisk == true {
		return multistep.ActionContinue
	}

	ui.Say("Resizing hard drive...")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepResizeDisk) buildResizeCommand(path string) []string {
	command := []string{"resize", "-f", s.Format}

	// add user-provided convert args
	command = append(command, s.QemuImgArgs.Resize...)

	// Add file and size
	command = append(command, path, s.DiskSize)

	return command
}

func (s *stepResizeDisk) Cleanup(state multistep.StateBag) {}
