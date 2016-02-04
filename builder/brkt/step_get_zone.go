package brkt

import (
	"fmt"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepGetZone struct {
	ComputingCell string
	Zone          string
}

func (s *stepGetZone) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*brkt.API)

	ui.Say("Getting Zone...")

	// machine type already set, go on
	if s.Zone != "" {
		ui.Say("Zone UUID supplied, using it...")

		state.Put("zone", s.Zone)
		return multistep.ActionContinue
	}

	zone, err := api.GetDefaultZone(s.ComputingCell)

	if err != nil {
		state.Put("error", fmt.Errorf("error while getting default Zone: %s", err))
		return multistep.ActionHalt
	}

	s.Zone = zone.Data.Id
	state.Put("zone", s.Zone)

	ui.Say("Found default Zone...")

	if s.Zone == "" {
		state.Put("error", fmt.Errorf("no zone set"))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepGetZone) Cleanup(state multistep.StateBag) {}
