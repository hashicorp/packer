package cvm

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepPreValidate struct {
	DestImageName string
	ForceDelete   bool
}

func (s *stepPreValidate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("TencentCloud support images with same name, image_name check is not required.")
	return multistep.ActionContinue
}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {}
