package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	vhdDir string = "Virtual Hard Disks"
	vmDir  string = "Virtual Machines"
)

type StepExportVm struct {
	OutputDir      string
	SkipCompaction bool
}

func (s *StepExportVm) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
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

	err = driver.ExportVirtualMachine(vmName, vmExportPath)
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
		err = driver.CompactDisks(expPath, vhdDir)
		if err != nil {
			errorMsg = "Error compacting disks: %s"
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Copying to output dir...")
	err = driver.CopyExportedVirtualMachine(expPath, outputPath, vhdDir, vmDir)
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
