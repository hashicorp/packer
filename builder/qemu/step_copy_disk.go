package qemu

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step copies the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCopyDisk struct{}

func (s *stepCopyDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packer.Ui)
	path := filepath.Join(config.OutputDir, fmt.Sprintf("%s.%s", config.VMName,
		strings.ToLower(config.Format)))
	name := config.VMName + "." + strings.ToLower(config.Format)

	command := []string{
		"convert",
		"-f", config.Format,
		isoPath,
		path,
	}

	if config.DiskImage == false {
		return multistep.ActionContinue
	}

	ui.Say("Copying hard drive...")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("disk_filename", name)

	return multistep.ActionContinue
}

func (s *stepCopyDisk) Cleanup(state multistep.StateBag) {}
