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

type StepStartVm struct {
}

func (s *StepStartVm) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error starting vm: %s"
	vmName := state.Get("vmName").(string)

	ui.Say("Starting vm...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Start-VM '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("'}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepStartVm) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)
	ui.Say("Stopping virtual machine...")

	var err error = nil

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Stop-VM '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Force}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf("Error stopping virtual machine: %s", err))
	}
}
