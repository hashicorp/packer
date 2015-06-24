// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
	"strconv"
	"time"
	powershell "github.com/mitchellh/packer/powershell"
)

const (
	SleepSeconds = 10
)

type StepWaitForPowerOff struct {
}

func (s *StepWaitForPowerOff) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)
	ui.Say("Waiting for vm to be powered down...")

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VM -Name $vmName).State -eq [Microsoft.HyperV.PowerShell.VMState]::Off")
	isOffScript := script.String()

	for {
		powershell := new(powershell.PowerShellCmd)
		cmdOut, err := powershell.Output(isOffScript, vmName);
		if err != nil {
			err := fmt.Errorf("Error checking VM's state: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if cmdOut == "True" {
			break
		} else {
			time.Sleep(time.Second * SleepSeconds);
		}
	}

	return multistep.ActionContinue
}

func (s *StepWaitForPowerOff) Cleanup(state multistep.StateBag) {
}

type StepWaitForInstallToComplete struct {
	ExpectedRebootCount uint
	ActionName string
}

func (s *StepWaitForInstallToComplete) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if(len(s.ActionName)>0){
		ui.Say(fmt.Sprintf("%v ! Waiting for VM to reboot %v times...",s.ActionName, s.ExpectedRebootCount))
	}

	var rebootCount uint
	var lastUptime uint64

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VM -Name $vmName).Uptime.TotalSeconds")

	uptimeScript := script.String()

	for rebootCount < s.ExpectedRebootCount {
		powershell := new(powershell.PowerShellCmd)
		cmdOut, err := powershell.Output(uptimeScript, vmName);
		if err != nil {
			err := fmt.Errorf("Error checking uptime: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		uptime, _ := strconv.ParseUint(strings.TrimSpace(string(cmdOut)), 10, 64)
		if uint64(uptime) < lastUptime {
			rebootCount++
			ui.Say(fmt.Sprintf("%v  -> Detected reboot %v after %v seconds...", s.ActionName, rebootCount, lastUptime))
		}

		lastUptime = uptime

		if (rebootCount < s.ExpectedRebootCount) {
			time.Sleep(time.Second * SleepSeconds);
		}
	}


	return multistep.ActionContinue
}

func (s *StepWaitForInstallToComplete) Cleanup(state multistep.StateBag) {

}
