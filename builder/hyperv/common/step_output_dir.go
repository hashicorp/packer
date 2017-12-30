package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// StepOutputDir sets up the output directory by creating it if it does
// not exist, deleting it if it does exist and we're forcing, and cleaning
// it up when we're done with it.
type StepOutputDir struct {
	Force bool
	Path  string

	cleanup bool
}

func (s *StepOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if _, err := os.Stat(s.Path); err == nil {
		if !s.Force {
			err := fmt.Errorf(
				"Output directory exists: %s\n\n"+
					"Use the force flag to delete it prior to building.",
				s.Path)
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Say("Deleting previous output directory...")
		os.RemoveAll(s.Path)
	}

	// Enable cleanup
	s.cleanup = true

	// Create the directory
	if err := os.MkdirAll(s.Path, 0755); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Make sure we can write in the directory
	f, err := os.Create(filepath.Join(s.Path, "_packer_perm_check"))
	if err != nil {
		err = fmt.Errorf("Couldn't write to output directory: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}
	f.Close()
	os.Remove(f.Name())

	return multistep.ActionContinue
}

func (s *StepOutputDir) Cleanup(state multistep.StateBag) {
	if !s.cleanup {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		ui := state.Get("ui").(packer.Ui)

		ui.Say("Deleting output directory...")
		for i := 0; i < 5; i++ {
			err := os.RemoveAll(s.Path)
			if err == nil {
				break
			}

			log.Printf("Error removing output dir: %s", err)
			time.Sleep(2 * time.Second)
		}
	}
}
