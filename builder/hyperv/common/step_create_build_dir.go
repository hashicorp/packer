package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateBuildDir struct {
	// User supplied directory under which we create the main build
	// directory. The build directory is used to house the VM files and
	// folders during the build. If unspecified the default temp directory
	// for the OS is used
	TempPath string
	// The full path to the build directory. This is the concatenation of
	// TempPath plus a directory uniquely named for the build
	dirPath string
}

// Creates the main directory used to house the VMs files and folders
// during the build
func (s *StepCreateBuildDir) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating build directory...")

	if s.TempPath == "" {
		s.TempPath = os.TempDir()
	}

	packerTempDir, err := ioutil.TempDir(s.TempPath, "packerhv")
	if err != nil {
		err := fmt.Errorf("Error creating build directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.dirPath = packerTempDir
	state.Put("packerTempDir", packerTempDir)

	return multistep.ActionContinue
}

// Cleanup removes the build directory
func (s *StepCreateBuildDir) Cleanup(state multistep.StateBag) {
	if s.dirPath == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Deleting build directory...")

	err := os.RemoveAll(s.dirPath)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting build directory: %s", err))
	}
}
