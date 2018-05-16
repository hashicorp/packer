package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"fmt"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"time"
	"context"
)

type StepWaitForIp struct{}

func (s *StepWaitForIp) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Waiting for IP...")

	ipChan := make(chan string)
	errChan := make(chan error)
	go func() {
		ip, err := vm.WaitForIP(ctx)
		if err != nil {
			errChan <- err
		} else {
			ipChan <- ip
		}
	}()

	for {
		select {
		case err := <-errChan:
			state.Put("error", err)
			return multistep.ActionHalt
		case <-ctx.Done():
			return multistep.ActionHalt
		case ip := <-ipChan:
			state.Put("ip", ip)
			ui.Say(fmt.Sprintf("IP address: %v", ip))
			return multistep.ActionContinue
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				return multistep.ActionHalt
			}
		}
	}
}

func (s *StepWaitForIp) Cleanup(state multistep.StateBag) {}
