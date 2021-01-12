package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This step starts the virtual machine.
//
// Uses:
//   driver Driver
//   ui packersdk.Ui
//   vmName string
//
// Produces:
type StepRun struct {
	Headless bool

	vmName string
}

func (s *StepRun) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Starting the virtual machine...")
	guiArgument := "gui"
	if s.Headless {
		vrdpIpRaw, vrdpIpOk := state.GetOk("vrdpIp")
		vrdpPortRaw, vrdpPortOk := state.GetOk("vrdpPort")

		if vrdpIpOk && vrdpPortOk {
			vrdpIp := vrdpIpRaw.(string)
			vrdpPort := vrdpPortRaw.(int)

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
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", s.vmName)

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.VBoxManage("controlvm", s.vmName, "poweroff"); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}
