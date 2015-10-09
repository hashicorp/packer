package docker

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"path/filepath"
)

// StepTempDir creates a temporary directory that we use in order to
// share data with the docker container over the communicator.
type StepTempDir struct {
	tempDir string
}

func (s *StepTempDir) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating a temporary directory for sharing data...")
	// Create the docker temp files in the current working directory
	// to work around an issue when running with docker-machine
	// using vm's needing access to shared folder content. This assumes
	// the current working directory is mapped as a share folder.
	// Allow TMPDIR to override this location.
	path := ""
	if tmpdir := os.Getenv("TMPDIR"); tmpdir == "" {
		abspath, err := filepath.Abs(".")
		if err == nil {
			path = abspath
		}
	}
	td, err := ioutil.TempDir(path, "packer-docker")
	if err != nil {
		err := fmt.Errorf("Error making temp dir: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.tempDir = td
	state.Put("temp_dir", s.tempDir)
	return multistep.ActionContinue
}

func (s *StepTempDir) Cleanup(state multistep.StateBag) {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}
