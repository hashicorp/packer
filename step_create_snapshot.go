package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"context"
	"github.com/vmware/govmomi/object"
)

type StepCreateSnapshot struct{
	createSnapshot bool
}

func (s *StepCreateSnapshot) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*object.VirtualMachine)
	ctx := state.Get("ctx").(context.Context)

	if s.createSnapshot {
		ui.Say("creating snapshot...")

		_, err := vm.CreateSnapshot(ctx, "packer_snapshot", "", true, true)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		ui.Say("done")
	}

	return multistep.ActionContinue
}

func (s *StepCreateSnapshot) Cleanup(state multistep.StateBag) {}
