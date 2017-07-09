package vsphere_tpl

import (
	"context"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/find"
)

type StepFetchVm struct {
	VMName string
}

func (s *StepFetchVm) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ctx := state.Get("context").(context.Context)
	f := state.Get("finder").(*find.Finder)

	ui.Say("Fetching VM...")

	vm, err := f.VirtualMachine(ctx, s.VMName)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vm", vm)
	return multistep.ActionContinue
}

func (s *StepFetchVm) Cleanup(multistep.StateBag) {}
