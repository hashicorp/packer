package docker

import (
	"context"
	"fmt"
	"log"
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
	configdir, err := packer.ConfigDir()
	if err != nil {
		return "", err
	}
	if tmpdir := os.Getenv("PACKER_TMP_DIR"); tmpdir != "" {
		// override the config dir with tmp dir. Still stat it and mkdirall if
		// necessary.
		fp, err := filepath.Abs(tmpdir)
		log.Printf("found PACKER_TMP_DIR env variable; setting tmpdir to %s", fp)
		if err != nil {
			return "", err
		}
		configdir = fp
	}

	td := filepath.Join(configdir, "tmp")
	_, err = os.Stat(td)
	if os.IsNotExist(err) {
		log.Printf("Creating tempdir in %s", td)
		if err = os.MkdirAll(td, 0755); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	log.Printf("Set Packer temp dir to %s", td)
	return td, nil
}

func (s *StepTempDir) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
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
