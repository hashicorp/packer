package qemu

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"os"
)

// This step shrinks the virtual disk that was used as the
// hard drive for the virtual machine.
type stepShrinkDisk struct{}

func (s *stepShrinkDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	sourcePath := state.Get("disk_filename")
	ui := state.Get("ui").(packer.Ui)
	name := config.VMName + ".shrink." + strings.ToLower(config.Format)
	targetPath := filepath.Join(config.OutputDir, name)

	command := []string{
		"convert",
		"-f", config.Format,
		sourcePath,
		targetPath,
	}

	if config.ShrinkImage == false {
		return multistep.ActionContinue
	}

	ui.Say("Shrinking hard drive...")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error shrinking hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if _, err := os.Stat(targetPath); err == nil && config.PackerForce {
		ui.Say("Deleting unshrinked disk image...")
		os.RemoveAll(targetPath)
	}

	state.Put("disk_filename", name)

	return multistep.ActionContinue
}

func (s *stepShrinkDisk) Cleanup(state multistep.StateBag) {}
