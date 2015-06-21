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
	"strings"
	"log"
)

type StepExecuteOnlineActivation struct {
}

func (s *StepExecuteOnlineActivation) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)

	errorMsg := "Error Executing Online Activation: %s"

	var remoteCmd packer.RemoteCmd
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	var err error

	ui.Say("Executing Online Activation...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("{ cscript \"$env:SystemRoot/system32/slmgr.vbs\" -ato //nologo }")

	remoteCmd.Command = "-ScriptBlock " + blockBuffer.String()

	remoteCmd.Stdout = stdout
	remoteCmd.Stderr = stderr

	err = comm.Start(&remoteCmd)

	stderrString := strings.TrimSpace(stderr.String())
	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	if len(stderrString) > 0 {
		err = fmt.Errorf(errorMsg, stderrString)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(stdoutString)

	return multistep.ActionContinue
}

func (s *StepExecuteOnlineActivation) Cleanup(state multistep.StateBag) {
	// do nothing
}
