package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
	"context"
)

type StepConvertToTemplate struct{
	ConvertToTemplate bool
}

func (s *StepConvertToTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*object.VirtualMachine)
	ctx := state.Get("ctx").(context.Context)

	// Turning into template if needed
	if s.ConvertToTemplate {
		ui.Say("turning into template...")
		err := vm.MarkAsTemplate(ctx)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		ui.Say("done")
	}

	return multistep.ActionContinue
}

func (s *StepConvertToTemplate) Cleanup(state multistep.StateBag) {}
