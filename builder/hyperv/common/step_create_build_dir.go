package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
)

type StepCreateBuildDir struct {
	// User supplied directory under which we create the main build
	// directory. The build directory is used to house the VM files and
	// folders during the build. If unspecified the default temp directory
	// for the OS is used
	TempPath string
	// The full path to the build directory. This is the concatenation of
	// TempPath plus a directory uniquely named for the build
	buildDir string
}

// Creates the main directory used to house the VMs files and folders
// during the build
func (s *StepCreateBuildDir) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Creating build directory...")

	var err error
	if s.TempPath == "" {
		s.buildDir, err = tmp.Dir("hyperv")
	} else {
		s.buildDir, err = ioutil.TempDir(s.TempPath, "hyperv")
	}

	if err != nil {
		err = fmt.Errorf("Error creating build directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Created build directory: %s", s.buildDir)

	// Record the build directory location for later steps
	state.Put("build_dir", s.buildDir)

	return multistep.ActionContinue
}

// Cleanup removes the build directory
func (s *StepCreateBuildDir) Cleanup(state multistep.StateBag) {
	if s.buildDir == "" {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Deleting build directory...")

	err := os.RemoveAll(s.buildDir)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting build directory: %s", err))
	}
}
