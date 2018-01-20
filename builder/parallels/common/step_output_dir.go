package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepOutputDir sets up the output directory by creating it if it does
// not exist, deleting it if it does exist and we're forcing, and cleaning
// it up when we're done with it.
type StepOutputDir struct {
	Force   bool
	Path    string
	success bool
}

// Run sets up the output directory.
func (s *StepOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if _, err := os.Stat(s.Path); err == nil && s.Force {
		ui.Say("Deleting previous output directory...")
		os.RemoveAll(s.Path)
	}

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

	s.success = true
	return multistep.ActionContinue
}

// Cleanup deletes the output directory.
func (s *StepOutputDir) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !s.success {
		return
	}

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
