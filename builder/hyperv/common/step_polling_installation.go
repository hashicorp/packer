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
//	"net"
	"log"
	"os/exec"
	"strings"
	"bytes"
)

const port string = "13000"

type StepPollingInstalation struct {
	step int
}

func (s *StepPollingInstalation) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error polling VM: %s"
	vmIp := state.Get("ip").(string)

	ui.Say("Start polling VM to check the installation is complete...")
/*
	count := 30
	var minutes time.Duration = 1
	sleepMin := time.Minute * minutes
	host := vmIp + ":" + port

	timeoutSec := time.Second * 15

	for count > 0 {
		ui.Say(fmt.Sprintf("Connecting vm (%s)...", host ))
		conn, err := net.DialTimeout("tcp", host, timeoutSec)
		if err == nil {
			ui.Say("Done!")
			conn.Close()
			break;
		}

		log.Println(err)
		ui.Say(fmt.Sprintf("Waiting more %v minutes...", uint(minutes)))
		time.Sleep(sleepMin)
		count--
	}

	if count == 0 {
		err := fmt.Errorf(errorMsg, "a signal from vm was not received in a given time period ")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
*/
	host := "'" + vmIp + "'," + port

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {function foo(){try{$client=New-Object System.Net.Sockets.TcpClient(")
	blockBuffer.WriteString(host)
	blockBuffer.WriteString(") -ErrorAction SilentlyContinue;if($client -eq $null){return $false}}catch{return $false}return $true} foo}")

	count := 60
	var duration time.Duration = 20
	sleepTime := time.Second * duration

	var res string

	for count > 0 {
		log.Println(fmt.Sprintf("Connecting vm (%s)...", host ))
		cmd := exec.Command("powershell", blockBuffer.String())
		cmdOut, err := cmd.Output()
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		res = strings.TrimSpace(string(cmdOut))

		if res != "False" {
			ui.Say("Signal was received from the VM")
			// Sleep before starting provision
			time.Sleep(time.Second*30)
			break;
		}

		log.Println(fmt.Sprintf("Slipping for more %v seconds...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if count == 0 {
		err := fmt.Errorf(errorMsg, "a signal from vm was not received in a given time period ")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("The installation complete")

	return multistep.ActionContinue
}

func (s *StepPollingInstalation) Cleanup(state multistep.StateBag) {

}
