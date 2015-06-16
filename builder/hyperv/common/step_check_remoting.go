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
	"time"
	"log"
	"strings"
)

type StepCheckRemoting struct {
}

func (s *StepCheckRemoting) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)

	var err error
	errorMsg := "Error step CheckRemoting: %s"

	// check the remote connection is ready
	{
		var cmd packer.RemoteCmd
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		magicWord := "ready"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("{ Write-Host '"+ magicWord +"' }")

		cmd.Command = "-ScriptBlock " + blockBuffer.String()
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		count := 5
		var duration time.Duration = 1
		sleepTime := time.Minute * duration

		ui.Say("Checking PS remoting is ready...")

		for count > 0 {
			err = comm.Start(&cmd)
			if err != nil {
				err := fmt.Errorf(errorMsg, "Remote connection failed")
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			stderrString := strings.TrimSpace(stderr.String())
			stdoutString := strings.TrimSpace(stdout.String())

			log.Printf("stdout: %s", stdoutString)
			log.Printf("stderr: %s", stderrString)

			if stdoutString == magicWord {
				break;
			}

			log.Println(fmt.Sprintf("Waiting %v minutes for the remote connection to get ready...", uint(duration)))
			time.Sleep(sleepTime)
			count--
		}

		if count == 0 {
			err := fmt.Errorf(errorMsg, "Remote connection failed")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCheckRemoting) Cleanup(state multistep.StateBag) {
	// do nothing
}
