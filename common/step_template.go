package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"context"
)

type StepConvertToTemplate struct{
	ConvertToTemplate bool
}

func (s *StepConvertToTemplate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if s.ConvertToTemplate {
		ui.Say("Convert VM into template...")
		err := vm.ConvertToTemplate()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConvertToTemplate) Cleanup(state multistep.StateBag) {}
