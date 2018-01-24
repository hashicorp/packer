package common

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type StepRun struct {
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Power on VM...")

	err := vm.PowerOn()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Waiting for IP...")
	ip, err := vm.WaitForIP()
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
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Power off VM...")

	err := vm.PowerOff()
	if err != nil {
		ui.Error(err.Error())
	}
}
