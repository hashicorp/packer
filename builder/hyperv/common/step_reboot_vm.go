package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepRebootVm struct {
}

func (s *StepRebootVm) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	errorMsg := "Error rebooting vm: %s"
	vmName := state.Get("vmName").(string)

	ui.Say("Rebooting vm...")

	err := driver.RestartVirtualMachine(vmName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting the VM to complete rebooting (2 minutes)...")

	sleepTime := time.Minute * 2
	time.Sleep(sleepTime)

	return multistep.ActionContinue
}

func (s *StepRebootVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
