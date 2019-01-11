package vagrant

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepUp struct {
	TeardownMethod string
	Provider       string
}

func (s *StepUp) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Calling Vagrant Up...")

	var args []string
	if s.Provider != "" {
		args = append(args, fmt.Sprintf("--provider=%s", s.Provider))
	}

	_, _, err := driver.Up(args)

	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUp) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("%sing Vagrant box...", s.TeardownMethod))

	var err error
	if s.TeardownMethod == "halt" {
		err = driver.Halt()
	} else if s.TeardownMethod == "suspend" {
		err = driver.Suspend()
	} else if s.TeardownMethod == "destroy" {
		err = driver.Destroy()
	} else {
		// Should never get here because of template validation
		state.Put("error", fmt.Errorf("Invalid teardown method selected; must be either halt, suspend, or destory."))
	}
	if err != nil {
		state.Put("error", fmt.Errorf("Error halting Vagrant machine; please try to do this manually"))
	}
}
