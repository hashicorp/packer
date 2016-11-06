// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
)

type StepCreateTempDir struct {
	dirPath string
}

func (s *StepCreateTempDir) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary directory...")

	tempDir := os.TempDir()
	packerTempDir, err := ioutil.TempDir(tempDir, "packerhv")
	if err != nil {
		err := fmt.Errorf("Error creating temporary directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.dirPath = packerTempDir
	state.Put("packerTempDir", packerTempDir)

	//	ui.Say("packerTempDir = '" + packerTempDir + "'")

	return multistep.ActionContinue
}

func (s *StepCreateTempDir) Cleanup(state multistep.StateBag) {
	if s.dirPath == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary directory...")

	err := os.RemoveAll(s.dirPath)

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting temporary directory: %s", err))
	}
}
