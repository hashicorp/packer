package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepTempDir creates a temporary directory that we use in order to
// share data with the docker container over the communicator.
type StepTempDir struct {
	tempDir string
}

// ConfigTmpDir returns the configuration tmp directory for Docker
func ConfigTmpDir() (string, error) {
	if tmpdir := os.Getenv("PACKER_TMP_DIR"); tmpdir != "" {
		return filepath.Abs(tmpdir)
	}
	configdir, err := packer.ConfigDir()
	if err != nil {
		return "", err
	}
	td := filepath.Join(configdir, "tmp")
	_, err = os.Stat(td)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(td, 0755); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return td, nil
}

func (s *StepTempDir) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating a temporary directory for sharing data...")

	tempdir, err := ConfigTmpDir()
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
