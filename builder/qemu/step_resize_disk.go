package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"strings"
)

// This step resizes the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepResizeDisk struct{}

func (s *stepResizeDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	path := filepath.Join(config.OutputDir, fmt.Sprintf("%s.%s", config.VMName,
		strings.ToLower(config.Format)))

	command := []string{
		"resize",
		path,
		fmt.Sprintf("%vM", config.DiskSize),
	}

	if config.DiskImage == false {
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

func (s *stepResizeDisk) Cleanup(state multistep.StateBag) {}
