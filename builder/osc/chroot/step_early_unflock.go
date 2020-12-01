package chroot

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepEarlyUnflock unlocks the flock.
type StepEarlyUnflock struct{}

func (s *StepEarlyUnflock) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	cleanup := state.Get("flock_cleanup").(Cleanup)
	ui := state.Get("ui").(packersdk.Ui)

	log.Println("Unlocking file lock...")
	if err := cleanup.CleanupFunc(state); err != nil {
		err := fmt.Errorf("Error unlocking file lock: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepEarlyUnflock) Cleanup(state multistep.StateBag) {}
