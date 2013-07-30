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
//
// Produces:
//   copy_files_cleanup CleanupFunc - A function to clean up the copied files
//   early.
type StepCopyFiles struct {
	files []string
}

func (s *StepCopyFiles) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*Config)
	mountPath := state["mount_path"].(string)
	ui := state["ui"].(packer.Ui)

	s.files = make([]string, 0, len(config.CopyFiles))
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

			s.files = append(s.files, chrootPath)
		}
	}

	state["copy_files_cleanup"] = s.CleanupFunc
	return multistep.ActionContinue
}

func (s *StepCopyFiles) Cleanup(state map[string]interface{}) {
	ui := state["ui"].(packer.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepCopyFiles) CleanupFunc(map[string]interface{}) error {
	if s.files != nil {
		for _, file := range s.files {
			log.Printf("Removing: %s", file)
			if err := os.Remove(file); err != nil {
				return err
			}
		}
	}

	s.files = nil
	return nil
}

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
