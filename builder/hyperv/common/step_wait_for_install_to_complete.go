package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

const (
	SleepSeconds = 10
)

type StepWaitForPowerOff struct {
}

func (s *StepWaitForPowerOff) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)
	ui.Say("Waiting for vm to be powered down...")

	for {
		isOff, err := driver.IsOff(vmName)

		if err != nil {
			err := fmt.Errorf("Error checking if vm is off: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if isOff {
			break
		} else {
			time.Sleep(time.Second * SleepSeconds)
		}
	}

	return multistep.ActionContinue
}

func (s *StepWaitForPowerOff) Cleanup(state multistep.StateBag) {
}

type StepWaitForInstallToComplete struct {
	ExpectedRebootCount uint
	ActionName          string
}

func (s *StepWaitForInstallToComplete) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	if len(s.ActionName) > 0 {
		ui.Say(fmt.Sprintf("%v ! Waiting for VM to reboot %v times...", s.ActionName, s.ExpectedRebootCount))
	}

	var rebootCount uint
	var lastUptime uint64

	for rebootCount < s.ExpectedRebootCount {
		uptime, err := driver.Uptime(vmName)

		if err != nil {
			err := fmt.Errorf("Error checking uptime: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if uptime < lastUptime {
			rebootCount++
			ui.Say(fmt.Sprintf("%v  -> Detected reboot %v after %v seconds...", s.ActionName, rebootCount, lastUptime))
		}

		lastUptime = uptime

		if rebootCount < s.ExpectedRebootCount {
			time.Sleep(time.Second * SleepSeconds)
		}
	}

	return multistep.ActionContinue
}

func (s *StepWaitForInstallToComplete) Cleanup(state multistep.StateBag) {

}
