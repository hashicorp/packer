package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
	"fmt"
)

type StepRun struct {
	// TODO: add boot time to provide a proper timeout during cleanup
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	d := state.Get("driver").(Driver)
	vm := state.Get("vm").(*object.VirtualMachine)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Power on VM...")
	err := d.powerOn(vm)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Waiting for IP...")
	ip, err := d.WaitForIP(vm)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("ip", ip)
	ui.Say(fmt.Sprintf("IP address: %v", ip))

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		d := state.Get("driver").(Driver)
		vm := state.Get("vm").(*object.VirtualMachine)
		ui := state.Get("ui").(packer.Ui)

		ui.Say("Power off VM...")
		err := d.powerOff(vm)
		if err != nil {
			ui.Error(err.Error())
			return
		}
	}
}
