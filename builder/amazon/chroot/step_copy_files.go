package chroot

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"os"
	"path/filepath"
)

// StepCopyFiles copies some files from the host into the chroot environment.
type StepCopyFiles struct {
	mounts []string
}

func (s *StepCopyFiles) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*Config)
	mountPath := state["mount_path"].(string)
	ui := state["ui"].(packer.Ui)

	if len(config.CopyFiles) > 0 {
		ui.Say("Copying files from host to chroot...")
		for _, path := range config.CopyFiles {
			ui.Message(path)
			chrootPath := filepath.Join(mountPath, path)
			log.Printf("Copying '%s' to '%s'", path, chrootPath)

			if err := s.copySingle(chrootPath, path); err != nil {
				err := fmt.Errorf("Error copying file: %s", err)
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepCopyFiles) Cleanup(state map[string]interface{}) {}

func (s *StepCopyFiles) copySingle(dst, src string) error {
	// Stat the src file so we can copy the mode later
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Remove any existing destination file
	if err := os.Remove(dst); err != nil {
		return err
	}

	// Copy the files
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	if _, err := io.Copy(dstF, srcF); err != nil {
		return err
	}
	dstF.Close()

	// Match the mode
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}
