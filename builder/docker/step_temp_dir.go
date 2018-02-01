package docker

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepTempDir creates a temporary directory that we use in order to
// share data with the docker container over the communicator.
type StepTempDir struct {
	tempDir string
}

func (s *StepTempDir) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating a temporary directory for sharing data...")

	var err error
	var tempdir string

	configTmpDir, err := packer.ConfigTmpDir()
	if err == nil {
		tempdir, err = ioutil.TempDir(configTmpDir, "packer-docker")
	}
	if err != nil {
		err := fmt.Errorf("Error making temp dir: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.tempDir = tempdir
	state.Put("temp_dir", s.tempDir)
	return multistep.ActionContinue
}

func (s *StepTempDir) Cleanup(state multistep.StateBag) {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}
