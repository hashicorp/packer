// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step creates switch for VM.
//
// Produces:
//   SwitchName string - The name of the Switch
type StepCreateExternalSwitch struct {
	SwitchName    string
	oldSwitchName string
}

func (s *StepCreateExternalSwitch) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	vmName := state.Get("vmName").(string)
	errorMsg := "Error createing external switch: %s"
	var err error

	ui.Say("Creating external switch...")

	packerExternalSwitchName := "paes_" + uuid.New()

	err = driver.CreateExternalVirtualSwitch(vmName, packerExternalSwitchName)
	if err != nil {
		err := fmt.Errorf("Error creating switch: %s", err)
		state.Put(errorMsg, err)
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
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Unregistering and deleting external switch...")

	var err error = nil

	errMsg := "Error deleting external switch: %s"

	// connect the vm to the old switch
	if s.oldSwitchName == "" {
		ui.Error(fmt.Sprintf(errMsg, "the old switch name is empty"))
		return
	}

	err = driver.ConnectVirtualMachineNetworkAdapterToSwitch(vmName, s.oldSwitchName)
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
