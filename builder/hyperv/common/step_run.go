package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepRun struct {
	GuiCancelFunc context.CancelFunc
	Headless      bool
	vmName        string
}

func (s *StepRun) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Starting the virtual machine...")

	err := driver.Start(vmName)
	if err != nil {
		err := fmt.Errorf("Error starting vm: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	if !s.Headless {
		ui.Say("Attempting to connect with vmconnect...")
		s.GuiCancelFunc, err = driver.Connect(vmName)
		if err != nil {
			log.Printf(fmt.Sprintf("Non-fatal error starting vmconnect: %s. continuing...", err))
		}
	}
	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

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
