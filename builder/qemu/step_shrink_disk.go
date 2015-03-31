package qemu

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step shrinks the virtual disk that was used as the
// hard drive for the virtual machine.
type stepShrinkDisk struct{}

func (s *stepShrinkDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	sourcePath := state.Get("disk_filename")
	//isoPath := state.Get("iso_path").(string)
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

	if config.ShrinkImage == false {
		return multistep.ActionContinue
	}

	ui.Say("Shrinking hard drive...")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("disk_filename", name)

	return multistep.ActionContinue
}

func (s *stepShrinkDisk) Cleanup(state multistep.StateBag) {}
