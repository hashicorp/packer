package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepDisableVlan struct {
}

func (s *StepDisableVlan) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error disabling vlan: %s"
	vmName := state.Get("vmName").(string)
	switchName := state.Get("SwitchName").(string)

	ui.Say("Disabling vlan...")

	err := driver.UntagVirtualMachineNetworkAdapterVlan(vmName, switchName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepDisableVlan) Cleanup(state multistep.StateBag) {
	//do nothing
}
