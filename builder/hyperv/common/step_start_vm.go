// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
	"github.com/mitchellh/packer/powershell/hyperv"
)

type StepStartVm struct {
	Reason string
	StartUpDelay int
}

func (s *StepStartVm) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error starting vm: %s"
	vmName := state.Get("vmName").(string)

	ui.Say("Starting vm for " + s.Reason + "...")

	err := hyperv.StartVirtualMachine(vmName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if s.StartUpDelay != 0 {
		//sleepTime := s.StartUpDelay * time.Second
		sleepTime := 60 * time.Second

		ui.Say(fmt.Sprintf("   Waiting %v for vm to start...", sleepTime))
		time.Sleep(sleepTime);		
	}

	return multistep.ActionContinue
}

func (s *StepStartVm) Cleanup(state multistep.StateBag) {
}
