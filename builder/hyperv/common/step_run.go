package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepRun struct {
	GuiCancelFunc context.CancelFunc
	Headless      bool
	SwitchName    string
	vmName        string
}

func (s *StepRun) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Determine Host IP for HyperV machine...")
	hostIp, err := driver.GetHostAdapterIpAddressForSwitch(s.SwitchName)
	if err != nil {
		err := fmt.Errorf("Error getting host adapter ip address: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Host IP for the HyperV machine: %s", hostIp))
	state.Put("http_ip", hostIp)

	if !s.Headless {
		ui.Say("Attempting to connect with vmconnect...")
		s.GuiCancelFunc, err = driver.Connect(vmName)
		if err != nil {
			log.Printf(fmt.Sprintf("Non-fatal error starting vmconnect: %s. continuing...", err))
		}
	}

	ui.Say("Starting the virtual machine...")

	err = driver.Start(vmName)
	if err != nil {
		err := fmt.Errorf("Error starting vm: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if !s.Headless && s.GuiCancelFunc != nil {
		ui.Say("Disconnecting from vmconnect...")
		s.GuiCancelFunc()
	}

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.Stop(s.vmName); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}
