package common

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepCreateTempDir struct {
	TempPath    string
	VhdTempPath string
	dirPath     string
}

func (s *StepCreateTempDir) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary directory...")

	if s.TempPath == "" {
		s.TempPath = os.TempDir()
	}

	packerTempDir, err := ioutil.TempDir(s.TempPath, "packerhv")
	if err != nil {
		err := fmt.Errorf("Error creating temporary directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.dirPath = packerTempDir
	state.Put("packerTempDir", packerTempDir)

	if s.VhdTempPath == "" {
		// Fall back to regular temp dir if no separate VHD temp dir set.
		state.Put("packerVhdTempDir", packerTempDir)
	} else {
		packerVhdTempDir, err := ioutil.TempDir(s.VhdTempPath, "packerhv-vhd")
		if err != nil {
			err := fmt.Errorf("Error creating temporary VHD directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.dirPath = packerVhdTempDir
		state.Put("packerVhdTempDir", packerVhdTempDir)
	}

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
