// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/powershell/hyperv"
	"io/ioutil"
	"path/filepath"
)

const (
	vhdDir string = "Virtual Hard Disks"
	vmDir  string = "Virtual Machines"
)

type StepExportVm struct {
	OutputDir      string
	SkipCompaction bool
}

func (s *StepExportVm) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	var err error
	var errorMsg string

	vmName := state.Get("vmName").(string)
	tmpPath := state.Get("packerTempDir").(string)
	outputPath := s.OutputDir

	// create temp path to export vm
	errorMsg = "Error creating temp export path: %s"
	vmExportPath, err := ioutil.TempDir(tmpPath, "export")
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Exporting vm...")

	err = hyperv.ExportVirtualMachine(vmName, vmExportPath)
	if err != nil {
		errorMsg = "Error exporting vm: %s"
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// copy to output dir
	expPath := filepath.Join(vmExportPath, vmName)

	if s.SkipCompaction {
		ui.Say("Skipping disk compaction...")
	} else {
		ui.Say("Compacting disks...")
		err = hyperv.CompactDisks(expPath, vhdDir)
		if err != nil {
			errorMsg = "Error compacting disks: %s"
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Coping to output dir...")
	err = hyperv.CopyExportedVirtualMachine(expPath, outputPath, vhdDir, vmDir)
	if err != nil {
		errorMsg = "Error exporting vm: %s"
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepExportVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
