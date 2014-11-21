// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package lin

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type StepGeneralizeOs struct {
	Command string
}

func (s *StepGeneralizeOs) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)

	ui.Say("Executing OS generalization...")

	var stdout, stderr bytes.Buffer
	cmd := &packer.RemoteCmd{
		Command: s.Command,
		Stdout:  &stdout,
		Stderr:  &stderr,
	}

	if err := comm.Start(cmd); err != nil {
		err := fmt.Errorf("Failed executing OS generalization command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the command to run
	cmd.Wait()

	// If the command failed to run, notify the user in some way.
	if cmd.ExitStatus != 0 {
		state.Put("error", fmt.Errorf(
				"OS generalization has non-zero exit status.\n\nStdout: %s\n\nStderr: %s",
				stdout.String(), stderr.String()))
		return multistep.ActionHalt
	}

	log.Printf("OS generalization stdout: %s", stdout.String())
	log.Printf("OS generalization stderr: %s", stderr.String())

	return multistep.ActionContinue
}

func (s *StepGeneralizeOs) Cleanup(state multistep.StateBag) {
	// do nothing
}
