package iso

import (
	"fmt"
	"strconv"

	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	command := []string{
		"set", vmName,
		"--device-add", "hdd",
		"--type", config.DiskType,
		"--size", strconv.FormatUint(uint64(config.DiskSize), 10),
		"--iface", config.HardDriveInterface,
	}

	ui.Say("Creating hard drive...")
	err := driver.Prlctl(command...)
	if err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
