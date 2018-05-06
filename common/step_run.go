package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"strings"
	"context"
)

type RunConfig struct {
	BootOrder   string        `mapstructure:"boot_order"` // example: "floppy,cdrom,ethernet,disk"
}

func (c *RunConfig) Prepare() []error {
	var errs []error

	return errs
}

type StepRun struct {
	Config *RunConfig
}

func (s *StepRun) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Set boot order...")
	if s.Config.BootOrder != "" {
		if err := vm.SetBootOrder(strings.Split(s.Config.BootOrder, ",")); err != nil {
			state.Put("error", fmt.Errorf("error selecting boot order: %v", err))
			return multistep.ActionHalt
		}
	}

	ui.Say("Power on VM...")
	err := vm.PowerOn()
	if err != nil {
		state.Put("error", fmt.Errorf("error powering on VM: %v", err))
		return multistep.ActionHalt
	}

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
