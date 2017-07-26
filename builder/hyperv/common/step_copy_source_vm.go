package common

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// This step copies a source virtual machine into the execution path
type StepCopySourceVM struct {
	SourcePath string
	VMName     string
}

func (s *StepCopySourceVM) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Copying source virtual machine...")

	path := state.Get("packerTempDir").(string)

	err := driver.CopySourceVirtualMachine(s.SourcePath, s.VMName, path)
	if err != nil {
		err := fmt.Errorf("Error copying source virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCopySourceVM) Cleanup(state multistep.StateBag) {
}
