package qemu

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	name := config.VMName
	path := filepath.Join(config.OutputDir, name)

	command := []string{
		"create",
		"-f", config.Format,
	}

	if config.UseBackingFile {
		isoPath := state.Get("iso_path").(string)
		command = append(command, "-b", isoPath)
	}

	command = append(command,
		path,
		fmt.Sprintf("%vM", config.DiskSize),
	)

	if config.DiskImage && !config.UseBackingFile {
		return multistep.ActionContinue
	}

	ui.Say("Creating hard drive...")
	if err := driver.QemuImg(command...); err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("disk_filename", name)

	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
