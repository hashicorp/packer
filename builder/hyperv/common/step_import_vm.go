package common

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step imports the existing virtual machine.
//
// Produces:
//   VMName string - The name of the VM
type StepImportVM struct {
	VMName    string
	SourceDir string
}

func (s *StepImportVM) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Importing virtual machine...")

	path := state.Get("packerTempDir").(string)

	err := driver.ImportVirtualMachine(s.VMName, s.SourceDir, path)
	if err != nil {
		err := fmt.Errorf("Error importing virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.VMName)

	return multistep.ActionContinue
}

func (s *StepImportVM) Cleanup(state multistep.StateBag) {
	if s.VMName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Unregistering and deleting virtual machine...")

	err := driver.DeleteVirtualMachine(s.VMName)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
