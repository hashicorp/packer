package vagrantcloud

import (
	"github.com/mitchellh/multistep"
)

type stepPrepareUpload struct {
}

func (s *stepPrepareUpload) Run(state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (s *stepPrepareUpload) Cleanup(state multistep.StateBag) {
}
