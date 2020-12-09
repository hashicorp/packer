package vagrant

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepUp struct {
	TeardownMethod string
	Provider       string
	GlobalID       string
}

func (s *StepUp) generateArgs() []string {
	box := "source"
	if s.GlobalID != "" {
		box = s.GlobalID
	}

	// start only the source box
	args := []string{box}

	if s.Provider != "" {
		args = append(args, fmt.Sprintf("--provider=%s", s.Provider))
	}
	return args
}

func (s *StepUp) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Calling Vagrant Up (this can take some time)...")

	args := s.generateArgs()
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", args[0])
	_, _, err := driver.Up(args)

	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUp) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("%sing Vagrant box...", s.TeardownMethod))

	box := "source"
	if s.GlobalID != "" {
		box = s.GlobalID
	}

	var err error
	if s.TeardownMethod == "halt" {
		err = driver.Halt(box)
	} else if s.TeardownMethod == "suspend" {
		err = driver.Suspend(box)
	} else if s.TeardownMethod == "destroy" {
		err = driver.Destroy(box)
	} else {
		// Should never get here because of template validation
		state.Put("error", fmt.Errorf("Invalid teardown method selected; must be either halt, suspend, or destroy."))
	}
	if err != nil {
		state.Put("error", fmt.Errorf("Error halting Vagrant machine; please try to do this manually"))
	}
}
