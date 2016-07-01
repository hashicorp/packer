package qemu

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"os"
)

// This step converts the virtual disk that was used as the
// hard drive for the virtual machine.
type stepConvertDisk struct{}

func (s *stepConvertDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	diskName := state.Get("disk_filename").(string)
	ui := state.Get("ui").(packer.Ui)

	if config.SkipCompaction && !config.DiskCompression {
		return multistep.ActionContinue
	}

	name := diskName + ".convert"

	sourcePath := filepath.Join(config.OutputDir, diskName)
	targetPath := filepath.Join(config.OutputDir, name)

	command := []string{
		"convert",
	}

	if config.DiskCompression {
		command = append(command, "-c")
	}

	command = append(command, []string{
		"-f", config.Format,
		"-O", config.Format,
		sourcePath,
		targetPath,
	}...,
	)

	ui.Say("Converting hard drive...")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error converting hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := os.Rename(targetPath, sourcePath); err != nil {
		err := fmt.Errorf("Error moving converted hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepConvertDisk) Cleanup(state multistep.StateBag) {}
