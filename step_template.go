package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
)

type StepConvertToTemplate struct{
	ConvertToTemplate bool
}

func (s *StepConvertToTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*object.VirtualMachine)
	d := state.Get("driver").(Driver)

	// Turning into template if needed
	if s.ConvertToTemplate {
		ui.Say("turning into template...")
		err := vm.MarkAsTemplate(d.ctx)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		ui.Say("done")
	}

	return multistep.ActionContinue
}

func (s *StepConvertToTemplate) Cleanup(state multistep.StateBag) {}
