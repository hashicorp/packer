package common

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateTempDir struct {
	// The user-supplied root directores into which we create subdirectories.
	TempPath    string
	VhdTempPath string
	// The subdirectories with the randomly generated name.
	dirPath    string
	vhdDirPath string
}

func (s *StepCreateTempDir) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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

		s.vhdDirPath = packerVhdTempDir
		state.Put("packerVhdTempDir", packerVhdTempDir)
	}

	//	ui.Say("packerTempDir = '" + packerTempDir + "'")

	return multistep.ActionContinue
}

func (s *StepCreateTempDir) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	if s.dirPath != "" {
		ui.Say("Deleting temporary directory...")

		err := os.RemoveAll(s.dirPath)

		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting temporary directory: %s", err))
		}
	}

	if s.vhdDirPath != "" && s.dirPath != s.vhdDirPath {
		ui.Say("Deleting temporary VHD directory...")

		err := os.RemoveAll(s.vhdDirPath)

		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting temporary VHD directory: %s", err))
		}
	}
}
