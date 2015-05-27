package iso

import (
	"fmt"

	"github.com/mitchellh/multistep"
	parallelscommon "github.com/mitchellh/packer/builder/parallels/common"
	"github.com/mitchellh/packer/packer"
)

// This step creates the actual virtual machine.
//
// Produces:
//   vmName string - The name of the VM
type stepCreateVM struct {
	vmName string
}

func (s *stepCreateVM) Run(state multistep.StateBag) multistep.StepAction {

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	name := config.VMName

	commands := make([][]string, 8)
	commands[0] = []string{
		"create", name,
		"--distribution", config.GuestOSType,
		"--dst", config.OutputDir,
		"--vmtype", "vm",
		"--no-hdd",
	}
	commands[1] = []string{"set", name, "--cpus", "1"}
	commands[2] = []string{"set", name, "--memsize", "512"}
	commands[3] = []string{"set", name, "--startup-view", "same"}
	commands[4] = []string{"set", name, "--on-shutdown", "close"}
	commands[5] = []string{"set", name, "--on-window-close", "keep-running"}
	commands[6] = []string{"set", name, "--auto-share-camera", "off"}
	commands[7] = []string{"set", name, "--smart-guard", "off"}

	ui.Say("Creating virtual machine...")
	for _, command := range commands {
		err := driver.Prlctl(command...)
		ui.Say(fmt.Sprintf("Executing: prlctl %s", command))
		if err != nil {
			err := fmt.Errorf("Error creating VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set the VM name property on the first command
		if s.vmName == "" {
			s.vmName = name
		}
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
