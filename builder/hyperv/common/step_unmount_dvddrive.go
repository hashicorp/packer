// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepUnmountDvdDrive struct {
}

func (s *StepUnmountDvdDrive) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	vmName := state.Get("vmName").(string)

	ui.Say("Unmounting dvd drive...")

	err := driver.UnmountDvdDrive(vmName)
	if err != nil {
		err := fmt.Errorf("Error unmounting dvd drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUnmountDvdDrive) Cleanup(state multistep.StateBag) {
}
