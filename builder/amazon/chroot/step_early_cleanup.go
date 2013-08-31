package chroot

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// StepEarlyCleanup performs some of the cleanup steps early in order to
// prepare for snapshotting and creating an AMI.
type StepEarlyCleanup struct{}

func (s *StepEarlyCleanup) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	cleanupKeys := []string{
		"copy_files_cleanup",
		"mount_extra_cleanup",
		"mount_device_cleanup",
		"attach_cleanup",
	}

	for _, key := range cleanupKeys {
		c := state.Get(key).(Cleanup)
		log.Printf("Running cleanup func: %s", key)
		if err := c.CleanupFunc(state); err != nil {
			err := fmt.Errorf("Error cleaning up: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepEarlyCleanup) Cleanup(state multistep.StateBag) {}
