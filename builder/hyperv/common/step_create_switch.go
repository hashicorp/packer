package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

const (
	SwitchTypeInternal = "Internal"
	SwitchTypePrivate  = "Private"
	DefaultSwitchType  = SwitchTypeInternal
)

// This step creates switch for VM.
//
// Produces:
//   SwitchName string - The name of the Switch
type StepCreateSwitch struct {
	// Specifies the name of the switch to be created.
	SwitchName string
	// Specifies the type of the switch to be created. Allowed values are Internal and Private. To create an External
	// virtual switch, specify either the NetAdapterInterfaceDescription or the NetAdapterName parameter, which
	// implicitly set the type of the virtual switch to External.
	SwitchType string
	// Specifies the name of the network adapter to be bound to the switch to be created.
	NetAdapterName string
	// Specifies the interface description of the network adapter to be bound to the switch to be created.
	NetAdapterInterfaceDescription string

	createdSwitch bool
}

func (s *StepCreateSwitch) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if len(s.SwitchType) == 0 {
		s.SwitchType = DefaultSwitchType
	}

	ui.Say(fmt.Sprintf("Creating switch '%v' if required...", s.SwitchName))

	createdSwitch, err := driver.CreateVirtualSwitch(s.SwitchName, s.SwitchType)
	if err != nil {
		err := fmt.Errorf("Error creating switch: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		s.SwitchName = ""
		return multistep.ActionHalt
	}

	s.createdSwitch = createdSwitch

	if !s.createdSwitch {
		ui.Say(fmt.Sprintf("    switch '%v' already exists. Will not delete on cleanup...", s.SwitchName))
	}

	// Set the final name in the state bag so others can use it
	state.Put("SwitchName", s.SwitchName)

	return multistep.ActionContinue
}

func (s *StepCreateSwitch) Cleanup(state multistep.StateBag) {
	if len(s.SwitchName) == 0 || !s.createdSwitch {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Unregistering and deleting switch...")

	err := driver.DeleteVirtualSwitch(s.SwitchName)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting switch: %s", err))
	}
}
