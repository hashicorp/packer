package vsphere_template

import (
	"context"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
)

type StepMarkAsTemplate struct{}

func (s *StepMarkAsTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ctx := state.Get("context").(context.Context)
	vm := state.Get("vm").(*object.VirtualMachine)

	ui.Say("Marking as a template...")

	if err := vm.MarkAsTemplate(ctx); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *StepMarkAsTemplate) Cleanup(multistep.StateBag) {}
