package common

import (
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"time"
)

type StepSleep struct {
	Minutes    time.Duration
	ActionName string
}

func (s *StepSleep) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if len(s.ActionName) > 0 {
		ui.Say(s.ActionName + "! Waiting for " + fmt.Sprintf("%v", uint(s.Minutes)) + " minutes to let the action to complete...")
	}
	time.Sleep(time.Minute * s.Minutes)

	return multistep.ActionContinue
}

func (s *StepSleep) Cleanup(state multistep.StateBag) {

}
