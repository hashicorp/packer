package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step imports an existing virtual machine in the OVF format.
//
// Produces:
//   vmName string - The name of the VM
type stepImportVM struct {
	vmName string
}

func (s *stepImportVM) Run(state map[string]interface{}) multistep.StepAction {
	var commands [][]string
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	name := config.VMName

	commands = make([][]string, 1)
	commands[0] = []string{"import", config.SourceOVF, "--options", "keepallmacs", "--vsys", "0", "--vmname", name}

	ui.Say("importing virtual machine...")
	for _, command := range commands {
		err := driver.VBoxManage(command...)
		if err != nil {
			err := fmt.Errorf("Error importing VM: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set the VM name property on the first command
		if s.vmName == "" {
			s.vmName = name
		}
	}

	// Set the final name in the state bag so others can use it
	state["vmName"] = s.vmName

	return multistep.ActionContinue
}

func (s *stepImportVM) Cleanup(state map[string]interface{}) {
	if s.vmName == "" {
		return
	}

	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	ui.Say("Unregistering and deleting virtual machine...")
	if err := driver.VBoxManage("unregistervm", s.vmName, "--delete"); err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
