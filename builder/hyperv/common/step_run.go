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
)

type StepRun struct {
	BootWait time.Duration
	Headless bool

	vmName string
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Starting the virtual machine...")

	err := driver.Start(vmName)
	if err != nil {
		err := fmt.Errorf("Error starting vm: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	if int64(s.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", s.BootWait))
		wait := time.After(s.BootWait)
	WAITLOOP:
		for {
			select {
			case <-wait:
				break WAITLOOP
			case <-time.After(1 * time.Second):
				if _, ok := state.GetOk(multistep.StateCancelled); ok {
					return multistep.ActionHalt
				}
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.Stop(s.vmName); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}
