//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type RunConfig

package common

import (
	"context"
	"strings"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type RunConfig struct {
	// Priority of boot devices. Defaults to `disk,cdrom`
	BootOrder string `mapstructure:"boot_order"` // example: "floppy,cdrom,ethernet,disk"
}

type StepRun struct {
	Config   *RunConfig
	SetOrder bool
}

func (s *StepRun) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	if s.Config.BootOrder != "" {
		ui.Say("Set boot order...")
		order := strings.Split(s.Config.BootOrder, ",")
		if err := vm.SetBootOrder(order); err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	} else {
		if s.SetOrder {
			ui.Say("Set boot order temporary...")
			if err := vm.SetBootOrder([]string{"disk", "cdrom"}); err != nil {
				state.Put("error", err)
				return multistep.ActionHalt
			}
		}
	}

	ui.Say("Power on VM...")
	err := vm.PowerOn()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	if s.Config.BootOrder == "" && s.SetOrder {
		ui.Say("Clear boot order...")
		if err := vm.SetBootOrder([]string{"-"}); err != nil {
			state.Put("error", err)
			return
		}
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui.Say("Power off VM...")

	err := vm.PowerOff()
	if err != nil {
		ui.Error(err.Error())
	}
}
