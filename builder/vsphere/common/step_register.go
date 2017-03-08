package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepRegister struct {
	KeepRegistered bool
	Format         string
}

func (s *StepRegister) Run(state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (s *StepRegister) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (s.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with remote host (keep_registered = true)")
		return
	}

	ui.Say("Destroying virtual machine...")
	if err := driver.Destroy(); err != nil {
		ui.Error(fmt.Sprintf("Error destroying VM: %s", err))
	}
	// Wait for the machine to actually destroy
	for {
		destroyed, _ := driver.IsDestroyed()
		if destroyed {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
}
