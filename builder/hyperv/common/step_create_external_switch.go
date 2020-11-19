package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

// This step creates an external switch for the VM.
//
// Produces:
//   SwitchName string - The name of the Switch
type StepCreateExternalSwitch struct {
	SwitchName    string
	oldSwitchName string
}

// Run runs the step required to create an external switch. Depending on
// the connectivity of the host machine, the external switch will allow the
// build VM to connect to the outside world.
func (s *StepCreateExternalSwitch) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	vmName := state.Get("vmName").(string)
	errorMsg := "Error creating external switch: %s"
	var err error

	ui.Say("Creating external switch...")

	packerExternalSwitchName := "paes_" + uuid.TimeOrderedUUID()

	// CreateExternalVirtualSwitch checks for an existing external switch,
	// creating one if required, and connects the VM to it
	err = driver.CreateExternalVirtualSwitch(vmName, packerExternalSwitchName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		s.SwitchName = ""
		return multistep.ActionHalt
	}

	switchName, err := driver.GetVirtualMachineSwitchName(vmName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(switchName) == 0 {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", "Can't get the VM switch name")
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("External switch name is: '" + switchName + "'")

	if switchName != packerExternalSwitchName {
		s.SwitchName = ""
	} else {
		s.SwitchName = packerExternalSwitchName
		s.oldSwitchName = state.Get("SwitchName").(string)
	}

	// Set the final name in the state bag so others can use it
	state.Put("SwitchName", switchName)

	return multistep.ActionContinue
}

func (s *StepCreateExternalSwitch) Cleanup(state multistep.StateBag) {
	if s.SwitchName == "" {
		return
	}
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Unregistering and deleting external switch...")

	errMsg := "Error deleting external switch: %s"

	// connect the vm to the old switch
	if s.oldSwitchName == "" {
		ui.Error(fmt.Sprintf(errMsg, "the old switch name is empty"))
		return
	}

	err := driver.ConnectVirtualMachineNetworkAdapterToSwitch(vmName, s.oldSwitchName)
	if err != nil {
		ui.Error(fmt.Sprintf(errMsg, err))
		return
	}

	state.Put("SwitchName", s.oldSwitchName)

	err = driver.DeleteVirtualSwitch(s.SwitchName)
	if err != nil {
		ui.Error(fmt.Sprintf(errMsg, err))
	}
}
