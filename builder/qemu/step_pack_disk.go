package qemu

import (
	"os"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"strings"
)

// This step packs the image by removing unallocated disk space
type stepPackDisk struct{}

func (s *stepPackDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	path := filepath.Join(config.OutputDir, fmt.Sprintf("%s.%s", config.VMName,
		strings.ToLower(config.Format)))
    newpath := fmt.Sprintf("%v.conv", path)

	command := []string{
		"convert",
        "-c",
		"-f", config.Format,
		"-O", config.Format,
		path,
		newpath,
	}

	ui.Say("Packing image")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error packing image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := os.Remove(path); err != nil {
		err := fmt.Errorf("Error deleting old image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := os.Rename(newpath, path); err != nil {
		err := fmt.Errorf("Error renaming image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepPackDisk) Cleanup(state multistep.StateBag) {}
