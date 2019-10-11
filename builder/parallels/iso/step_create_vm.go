package iso

import (
	"context"
	"fmt"
	"strconv"

	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step creates the actual virtual machine.
//
// Produces:
//   vmName string - The name of the VM
type stepCreateVM struct {
	vmName string
}

func (s *stepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	name := config.VMName

	commands := make([][]string, 3)

	commands[0] = []string{
		"create", name,
		"--distribution", config.GuestOSType,
		"--dst", config.OutputDir,
		"--no-hdd",
	}
	commands[1] = []string{
		"set", name,
		"--cpus", strconv.Itoa(config.HWConfig.CpuCount),
	}
	commands[2] = []string{
		"set", name,
		"--memsize", strconv.Itoa(config.HWConfig.MemorySize),
	}

	if config.HWConfig.Sound {
		commands = append(commands, []string{
			"set", name,
			"--device-add-sound",
			"--connect",
		})
	}

	if config.HWConfig.USB {
		commands = append(commands, []string{
			"set", name,
			"--device-add-usb",
		})
	}

	ui.Say("Creating virtual machine...")
	for _, command := range commands {
		if err := driver.Prlctl(command...); err != nil {
			err := fmt.Errorf("Error creating VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Applying default settings...")
	if err := driver.SetDefaultConfiguration(name); err != nil {
		err := fmt.Errorf("Error VM configuration: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the VM name property on the first command
	if s.vmName == "" {
		s.vmName = name
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.vmName)
	return multistep.ActionContinue
}

func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering virtual machine...")
	if err := driver.Prlctl("unregister", s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering virtual machine: %s", err))
	}
}
