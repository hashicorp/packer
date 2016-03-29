package iso

import (
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/packer"
)

type StepRegister struct {
	registeredPath string
	Format         string
}

func (s *StepRegister) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	if remoteDriver, ok := driver.(RemoteDriver); ok {
		ui.Say("Registering remote VM...")
		if err := remoteDriver.Register(vmxPath); err != nil {
			err := fmt.Errorf("Error registering VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.registeredPath = vmxPath
	}

	return multistep.ActionContinue
}

func (s *StepRegister) Cleanup(state multistep.StateBag) {
	if s.registeredPath == "" {
		return
	}

	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (config.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with ESX host (keep_registered = true)")
		return
	}

	if remoteDriver, ok := driver.(RemoteDriver); ok {
		if s.Format == "" {
			ui.Say("Unregistering virtual machine...")
			if err := remoteDriver.Unregister(s.registeredPath); err != nil {
				ui.Error(fmt.Sprintf("Error unregistering VM: %s", err))
			}

			s.registeredPath = ""
		} else {
			ui.Say("Destroying virtual machine...")
			if err := remoteDriver.Destroy(); err != nil {
				ui.Error(fmt.Sprintf("Error destroying VM: %s", err))
			}
			// Wait for the machine to actually destroy
			for {
				destroyed, _ := remoteDriver.IsDestroyed()
				if destroyed {
					break
				}
				time.Sleep(150 * time.Millisecond)
			}
		}
	}
}
