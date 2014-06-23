package vagrantcloud

import (
	"github.com/mitchellh/multistep"
)

type stepUpload struct {
}

func (s *stepUpload) Run(state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (s *stepUpload) Cleanup(state multistep.StateBag) {
}
