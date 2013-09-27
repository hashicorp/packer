package chroot

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
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

func (s *StepCopyFiles) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)

	s.files = make([]string, 0, len(config.CopyFiles))
	if len(config.CopyFiles) > 0 {
		ui.Say("Copying files from host to chroot...")
		for _, path := range config.CopyFiles {
			ui.Message(path)
			chrootPath := filepath.Join(mountPath, path)
			log.Printf("Copying '%s' to '%s'", path, chrootPath)

			if err := copySingle(chrootPath, path, config.CopyCommand); err != nil {
				err := fmt.Errorf("Error copying file: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			s.files = append(s.files, chrootPath)
		}
	}

	state.Put("copy_files_cleanup", s)
	return multistep.ActionContinue
}

func (s *StepCopyFiles) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepCopyFiles) CleanupFunc(multistep.StateBag) error {
	if s.files != nil {
		for _, file := range s.files {
			log.Printf("Removing: %s", file)
			chrootCommand := fmt.Sprintf("rm %s", file)
			localCmd := exec.Command("/bin/sh", "-c", chrootCommand)
			if err := localCmd.Run(); err != nil {
				return err
			}
		}
	}

	s.files = nil
	return nil
}
