// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	//"fmt"
	"os"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/mitchellh/packer/powershell"
)

type StepUpdateIntegrationServices struct {
	Username string
	Password string

	newDvdDriveProperties dvdDriveProperties
}

type dvdDriveProperties struct {
	ControllerNumber string
	ControllerLocation string
}

func (s *StepUpdateIntegrationServices) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Mounting Integration Services Setup Disk...")

	_, err := s.mountIntegrationServicesSetupDisk(vmName);
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// dvdDriveLetter, err := s.getDvdDriveLetter(vmName)
	// if err != nil {
	// 	state.Put("error", err)
	// 	ui.Error(err.Error())
	// 	return multistep.ActionHalt
	// }

	// setup := dvdDriveLetter + ":\\support\\"+osArchitecture+"\\setup.exe /quiet /norestart"

	// ui.Say("Run: " + setup)

	return multistep.ActionContinue
}

func (s *StepUpdateIntegrationServices) Cleanup(state multistep.StateBag) {
	vmName := state.Get("vmName").(string)

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Set-VMDvdDrive -VMName $vmName -Path $null")

	powershell := new(powershell.PowerShellCmd)
	_ = powershell.Run(script.String(), vmName)
}

func (s *StepUpdateIntegrationServices) mountIntegrationServicesSetupDisk(vmName string) (dvdDriveProperties, error) {

	var dvdProperties dvdDriveProperties

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Add-VMDvdDrive -VMName $vmName")

	powershell := new(powershell.PowerShellCmd)
	err := powershell.Run(script.String(), vmName)
	if err != nil {
		return dvdProperties, err
	}

	script.Reset()
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VMDvdDrive -VMName $vmName | Where-Object {$_.Path -eq $null}).ControllerLocation")
	controllerLocation, err := powershell.Output(script.String(), vmName)
	if err != nil {
		return dvdProperties, err
	}

	script.Reset()
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VMDvdDrive -VMName $vmName | Where-Object {$_.Path -eq $null}).ControllerNumber")
	controllerNumber, err := powershell.Output(script.String(), vmName)
	if err != nil {
		return dvdProperties, err
	}

	isoPath := os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

	script.Reset()
	script.WriteLine("param([string]$vmName,[string]$path,[string]$controllerNumber,[string]$controllerLocation)")
	script.WriteLine("Set-VMDvdDrive -VMName $vmName -Path $path -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation")

	err = powershell.Run(script.String(), vmName, isoPath, controllerNumber, controllerLocation)
	if err != nil {
		return dvdProperties, err
	}

	dvdProperties.ControllerNumber = controllerNumber
	dvdProperties.ControllerLocation = controllerLocation

	return dvdProperties, err
}
