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
	"log"
)

type StepUnmountSecondaryDvdImages struct {
}

func (s *StepUnmountSecondaryDvdImages) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Unmounting Integration Services Setup Disk...")

	vmName := state.Get("vmName").(string)

	// todo: should this message say removing the dvd?

	dvdProperties := state.Get("secondary.dvd.properties").([]DvdControllerProperties)

	log.Println(fmt.Sprintf("Found DVD properties %s", len(dvdProperties)))

	for _, dvdProperty := range dvdProperties {
		controllerNumber := dvdProperty.ControllerNumber
		controllerLocation := dvdProperty.ControllerLocation

		var script powershell.ScriptBuilder
		powershell := new(powershell.PowerShellCmd)

		script.WriteLine("param([string]$vmName,[int]$controllerNumber,[int]$controllerLocation)")
		script.WriteLine("$vmDvdDrive = Get-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation")
		script.WriteLine("if (!$vmDvdDrive) {throw 'unable to find dvd drive'}")
		script.WriteLine("Remove-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation")
		err := powershell.Run(script.String(), vmName, controllerNumber, controllerLocation)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepUnmountSecondaryDvdImages) Cleanup(state multistep.StateBag) {
}
