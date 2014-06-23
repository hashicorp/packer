package vagrantcloud

import (
	"github.com/mitchellh/multistep"
)

type stepCreateProvider struct {
}

func (s *stepCreateProvider) Run(state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (s *stepCreateProvider) Cleanup(state multistep.StateBag) {
}
