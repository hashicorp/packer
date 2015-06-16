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

type StepEnableIntegrationService struct {
	name string
}

func (s *StepEnableIntegrationService) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	vmName := state.Get("vmName").(string)
	s.name = "Guest Service Interface"

	ui.Say("Enabling Integration Service...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Enable-VMIntegrationService -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Name '")
	blockBuffer.WriteString(s.name)
	blockBuffer.WriteString("'}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf("Error enabling Integration Service: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepEnableIntegrationService) Cleanup(state multistep.StateBag) {
	// do nothing
}
