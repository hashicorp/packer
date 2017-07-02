package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
	"fmt"
	"github.com/vmware/govmomi/vim25/types"
)

type StepRun struct{
	// TODO: add boot time to provide a proper timeout during cleanup
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*object.VirtualMachine)
	d := state.Get("driver").(Driver)

	ui.Say("VM power on...")
	task, err := vm.PowerOn(d.ctx)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	_, err = task.WaitForResult(d.ctx, nil)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("VM waiting for IP...")
	ip, err := vm.WaitForIP(d.ctx)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("ip", ip)
	ui.Say(fmt.Sprintf("VM ip %v", ip))
	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		vm := state.Get("vm").(*object.VirtualMachine)
		d := state.Get("driver").(Driver)
		ui := state.Get("ui").(packer.Ui)

		if state, err := vm.PowerState(d.ctx); state != types.VirtualMachinePowerStatePoweredOff && err == nil {
			ui.Say("shutting down VM...")

			task, err := vm.PowerOff(d.ctx)
			if err != nil {
				ui.Error(err.Error())
				return
			}
			_, err = task.WaitForResult(d.ctx, nil)
			if err != nil {
				ui.Error(err.Error())
				return
			}
		} else if err != nil {
			ui.Error(err.Error())
			return
		}
	}
}
