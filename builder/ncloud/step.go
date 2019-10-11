package ncloud

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func processStepResult(err error, sayError func(error), state multistep.StateBag) multistep.StepAction {
	if err != nil {
		state.Put("Error", err)
		sayError(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}
