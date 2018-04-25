package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"strings"
	"time"
	"context"
)

type RunConfig struct {
	BootOrder   string        `mapstructure:"boot_order"` // example: "floppy,cdrom,ethernet,disk"
	RawBootWait string        `mapstructure:"boot_wait"`  // example: "1m30s"; default: "10s"
	bootWait    time.Duration ``
}

func (c *RunConfig) Prepare() []error {
	var errs []error

	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}

	var err error
	c.bootWait, err = time.ParseDuration(c.RawBootWait)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed parsing boot_wait: %s", err))
	}

	return errs
}

type StepRun struct {
	Config *RunConfig
}

func (s *StepRun) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Power on VM...")

	if s.Config.BootOrder != "" {
		if err := vm.SetBootOrder(strings.Split(s.Config.BootOrder, ",")); err != nil {
			state.Put("error", fmt.Errorf("error selecting boot order: %v", err))
			return multistep.ActionHalt
		}
	}

	err := vm.PowerOn()
	if err != nil {
		state.Put("error", fmt.Errorf("error powering on VM: %v", err))
		return multistep.ActionHalt
	}

	if int64(s.Config.bootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", s.Config.bootWait))
		wait := time.After(s.Config.bootWait)
	WAITLOOP:
		for {
			select {
			case <-wait:
				break WAITLOOP
			case <-time.After(1 * time.Second):
				if _, ok := state.GetOk(multistep.StateCancelled); ok {
					return multistep.ActionHalt
				}
			}
		}
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
