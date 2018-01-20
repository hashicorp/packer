package common

import (
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepUnmountFloppyDrive struct {
	Generation uint
}

func (s *StepUnmountFloppyDrive) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if s.Generation > 1 {
		return multistep.ActionContinue
	}

	vmName := state.Get("vmName").(string)
	ui.Say("Unmount/delete floppy drive (Run)...")

	errorMsg := "Error Unmounting floppy drive: %s"

	err := driver.UnmountFloppyDrive(vmName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
	}

	return multistep.ActionContinue
}

func (s *StepUnmountFloppyDrive) Cleanup(state multistep.StateBag) {
	// do nothing
}
