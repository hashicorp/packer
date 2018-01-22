package common

import (
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step starts the virtual machine.
//
// Uses:
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepRun struct {
	BootWait time.Duration
	Headless bool

	vmName string
}

func (s *StepRun) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Starting the virtual machine...")
	guiArgument := "gui"
	if s.Headless {
		vrdpIpRaw, vrdpIpOk := state.GetOk("vrdpIp")
		vrdpPortRaw, vrdpPortOk := state.GetOk("vrdpPort")

		if vrdpIpOk && vrdpPortOk {
			vrdpIp := vrdpIpRaw.(string)
			vrdpPort := vrdpPortRaw.(uint)

			ui.Message(fmt.Sprintf(
				"The VM will be run headless, without a GUI. If you want to\n"+
					"view the screen of the VM, connect via VRDP without a password to\n"+
					"rdp://%s:%d", vrdpIp, vrdpPort))
		} else {
			ui.Message("The VM will be run headless, without a GUI, as configured.\n" +
				"If the run isn't succeeding as you expect, please enable the GUI\n" +
				"to inspect the progress of the build.")
		}
		guiArgument = "headless"
	}
	command := []string{"startvm", vmName, "--type", guiArgument}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	if int64(s.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", s.BootWait))
		wait := time.After(s.BootWait)
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
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.VBoxManage("controlvm", s.vmName, "poweroff"); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}
