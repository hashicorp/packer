package vagrantcloud

import (
	"github.com/mitchellh/multistep"
)

type stepVerifyUpload struct {
}

func (s *stepVerifyUpload) Run(state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (s *stepVerifyUpload) Cleanup(state multistep.StateBag) {
}
