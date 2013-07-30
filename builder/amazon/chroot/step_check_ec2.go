package chroot

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCheckEC2 struct{}

func (s *StepCheckEC2) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)
	ui.Say("Verifying we're on an EC2 instance...")
	return multistep.ActionContinue
}

func (s *StepCheckEC2) Cleanup(map[string]interface{}) {}
