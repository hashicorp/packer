// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step creates switch for VM.
//
// Produces:
//   SwitchName string - The name of the Switch
type StepCreateSwitch struct {
	SwitchName string
}

func (s *StepCreateSwitch) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating internal switch...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {$TestSwitch = Get-VMSwitch -Name '")
	blockBuffer.WriteString(s.SwitchName)
	blockBuffer.WriteString("' -ErrorAction SilentlyContinue; if ($TestSwitch.Count -EQ 0){New-VMSwitch -Name '")
	blockBuffer.WriteString(s.SwitchName)
	blockBuffer.WriteString("' -SwitchType Internal}}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf("Error creating switch: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		s.SwitchName = "";
		return multistep.ActionHalt
	}

	// Set the final name in the state bag so others can use it
	state.Put("SwitchName", s.SwitchName)

	return multistep.ActionContinue
}

func (s *StepCreateSwitch) Cleanup(state multistep.StateBag) {
	if s.SwitchName == "" {
		return
	}
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering and deleting switch...")

	var err error = nil

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Remove-VMSwitch '")
	blockBuffer.WriteString(s.SwitchName)
	blockBuffer.WriteString("' -Force}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting switch: %s", err))
	}
}
