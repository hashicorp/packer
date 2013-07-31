package chroot

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// StepEarlyUnflock unlocks the flock.
type StepEarlyUnflock struct{}

func (s *StepEarlyUnflock) Run(state map[string]interface{}) multistep.StepAction {
	cleanup := state["flock_cleanup"].(Cleanup)
	ui := state["ui"].(packer.Ui)

	log.Println("Unlocking file lock...")
	if err := cleanup.CleanupFunc(state); err != nil {
		err := fmt.Errorf("Error unlocking file lock: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepEarlyUnflock) Cleanup(state map[string]interface{}) {}
