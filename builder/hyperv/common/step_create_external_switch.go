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
	"os/exec"
	"strings"
	"code.google.com/p/go-uuid/uuid"
)

// This step creates switch for VM.
//
// Produces:
//   SwitchName string - The name of the Switch
type StepCreateExternalSwitch struct {
	SwitchName string
	oldSwitchName string
}

func (s *StepCreateExternalSwitch) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	vmName := state.Get("vmName").(string)
	errorMsg := "Error createing external switch: %s"
	var err error

	ui.Say("Creating external switch...")

	packerExternalSwitchName := "paes_" + uuid.New()

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$extSwitchName='"+ packerExternalSwitchName +"';")
	blockBuffer.WriteString("$vmName='"+ vmName +"';")
	blockBuffer.WriteString("$switch=$null; $names=@('ethernet','wi-fi','foo'); $adapters=foreach($name in $names){Get-NetAdapter -physical -Name $name -ErrorAction SilentlyContinue | where status -eq 'up' }foreach($adapter in $adapters){$switch=Get-VMSwitch –SwitchType External | where {$_.NetAdapterInterfaceDescription -eq $adapter.InterfaceDescription};if($switch -eq $null){$switch=New-VMSwitch -Name $extSwitchName -NetAdapterName $adapter.Name -AllowManagementOS $true -Notes 'Parent OS, VMs, WiFi'};if($switch -ne $null){break}};if($switch -ne $null){Get-VMNetworkAdapter –VMName $vmName | Connect-VMNetworkAdapter -VMSwitch $switch } else{ Write-Error 'No internet adapters found'}")
	blockBuffer.WriteString("}")

	err = driver.HypervManage(blockBuffer.String())

	if err != nil {
		err := fmt.Errorf("Error creating switch: %s", err)
		state.Put(errorMsg, err)
		ui.Error(err.Error())
		s.SwitchName = "";
		return multistep.ActionHalt
	}

	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {$adapter=Get-VMNetworkAdapter -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("'; $adapter.SwitchName}")

	cmd := exec.Command("powershell", blockBuffer.String())
	cmdOut, err := cmd.Output()
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	switchName := strings.TrimSpace(string(cmdOut))

	if len(switchName) == 0 {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", "Can't get the VM switch name")
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("External switch name is: '" + switchName + "'")

	if(switchName != packerExternalSwitchName){
		s.SwitchName = ""
	} else {
		s.SwitchName =  packerExternalSwitchName
		s.oldSwitchName = state.Get("SwitchName").(string)
	}

	// Set the final name in the state bag so others can use it
	state.Put("SwitchName", switchName)

	return multistep.ActionContinue
}

func (s *StepCreateExternalSwitch) Cleanup(state multistep.StateBag) {
	if s.SwitchName == "" {
		return
	}
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Unregistering and deleting external switch...")

	var err error = nil

	errMsg := "Error deleting external switch: %s"

	// connect the vm to the old switch
	if s.oldSwitchName == "" {
		ui.Error(fmt.Sprintf(errMsg, "the old switch name is empty"))
		return
	}

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$sn='" + s.oldSwitchName + "';")
	blockBuffer.WriteString("$vmName='" + vmName + "';")
	blockBuffer.WriteString("Get-VMNetworkAdapter –VMName $vmName | Connect-VMNetworkAdapter –SwitchName $sn")
	blockBuffer.WriteString("}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf(errMsg, err))
		return
	}

	state.Put("SwitchName", s.oldSwitchName)

	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$sn='" + s.SwitchName + "';")
	blockBuffer.WriteString("$TestSwitch = Get-VMSwitch -Name $sn -ErrorAction SilentlyContinue;if($TestSwitch -NE $null){Remove-VMSwitch $sn -Force}")
	blockBuffer.WriteString("}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf(errMsg, err))
	}
}
