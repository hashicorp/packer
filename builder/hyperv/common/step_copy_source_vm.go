package common

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// This step copies a source vhd/vhdx into the execution path
type StepCopySourceDisk struct {
	SourcePath string
	VMName     string
}

func (s *StepCopySourceDisk) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Copying source disk...")

	path := state.Get("packerTempDir").(string)

	err := driver.CopySourceVirtualMachine(s.SourcePath, s.VMName, path)
	if err != nil {
		err := fmt.Errorf("Error copying source disk: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCopySourceDisk) Cleanup(state multistep.StateBag) {
}
