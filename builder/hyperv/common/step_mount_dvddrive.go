// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/mitchellh/packer/powershell"
)

type StepMountDvdDrive struct {
	RawSingleISOUrl string
	path            string
}

func (s *StepMountDvdDrive) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error mounting dvd drive: %s"
	vmName := state.Get("vmName").(string)
	isoPath := s.RawSingleISOUrl

	// Check that there is a virtual dvd drive
	var script powershell.ScriptBuilder
	powershell := new(powershell.PowerShellCmd)

	script.Reset()
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VMDvdDrive -VMName $vmName).ControllerNumber")
	controllerNumber, err := powershell.Output(script.String(), vmName)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if controllerNumber == "" {
		// Add a virtual dvd drive as there is none
		script.Reset()
		script.WriteLine("param([string]$vmName)")
		script.WriteLine("Add-VMDvdDrive -VMName $vmName")
		script.WriteLine("$dvdDrive = Get-VMDvdDrive -VMName $vmName | Select-Object -first 1")
		script.WriteLine("Set-VMFirmware -VMName $vmName -FirstBootDevice $dvdDrive")
		err = powershell.Run(script.String(), vmName)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Mounting dvd drive...")

	err = driver.MountDvdDrive(vmName, isoPath)
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

	driver := state.Get("driver").(Driver)

	errorMsg := "Error unmounting dvd drive: %s"

	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unmounting dvd drive...")

	err := driver.UnmountDvdDrive(vmName)
	if err != nil {
		ui.Error(fmt.Sprintf(errorMsg, err))
	}
}
