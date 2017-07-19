package vsphere_template

import (
	"context"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
)

type stepMarkAsTemplate struct{}

func (s *stepMarkAsTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*object.VirtualMachine)

	ui.Say("Marking as a template...")

	if err := vm.MarkAsTemplate(context.Background()); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepMarkAsTemplate) Cleanup(multistep.StateBag) {}
