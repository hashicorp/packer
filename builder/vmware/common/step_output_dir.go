package common

import (
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepOutputDir sets up the output directory by creating it if it does
// not exist, deleting it if it does exist and we're forcing, and cleaning
// it up when we're done with it.
type StepOutputDir struct {
	Force bool

	success bool
}

func (s *StepOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	dir := state.Get("dir").(OutputDir)
	ui := state.Get("ui").(packer.Ui)

	exists, err := dir.DirExists()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if exists {
		if s.Force {
			ui.Say("Deleting previous output directory...")
			dir.RemoveAll()
		} else {
			state.Put("error", fmt.Errorf(
				"Output directory '%s' already exists.", dir.String()))
			return multistep.ActionHalt
		}
	}

	if err := dir.MkdirAll(); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.success = true
	return multistep.ActionContinue
}

func (s *StepOutputDir) Cleanup(state multistep.StateBag) {
	if !s.success {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		dir := state.Get("dir").(OutputDir)
		ui := state.Get("ui").(packer.Ui)

		exists, _ := dir.DirExists()
		if exists {
			ui.Say("Deleting output directory...")
			for i := 0; i < 5; i++ {
				err := dir.RemoveAll()
				if err == nil {
					break
				}

				log.Printf("Error removing output dir: %s", err)
				time.Sleep(2 * time.Second)
			}
		}
	}
}
