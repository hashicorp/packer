package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepSleep struct {
	Minutes    time.Duration
	ActionName string
}

func (s *StepSleep) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if len(s.ActionName) > 0 {
		ui.Say(s.ActionName + "! Waiting for " + fmt.Sprintf("%v", uint(s.Minutes)) +
			" minutes to let the action to complete...")
	}
	time.Sleep(time.Minute * s.Minutes)

	return multistep.ActionContinue
}

func (s *StepSleep) Cleanup(state multistep.StateBag) {

}
