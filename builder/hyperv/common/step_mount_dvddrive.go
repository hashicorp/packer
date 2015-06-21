// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/powershell/hyperv"
)


type StepMountDvdDrive struct {
	RawSingleISOUrl string
	path string
}

func (s *StepMountDvdDrive) Run(state multistep.StateBag) multistep.StepAction {
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error mounting dvd drive: %s"
	vmName := state.Get("vmName").(string)
	isoPath := s.RawSingleISOUrl

	ui.Say("Mounting dvd drive...")

	err := hyperv.MountDvdDrive(vmName, isoPath)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.path = isoPath

	return multistep.ActionContinue
}

func (s *StepMountDvdDrive) Cleanup(state multistep.StateBag) {
	if s.path == "" {
		return
	}

	errorMsg := "Error unmounting dvd drive: %s"

	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unmounting dvd drive...")

	err := hyperv.UnmountDvdDrive(vmName)
	if err != nil {
		ui.Error(fmt.Sprintf(errorMsg, err))
	}
}
