package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepRegister struct {
	registeredPath string
	Format         string
	KeepRegistered bool
	SkipExport     bool
}

func (s *StepRegister) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

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

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (s.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with ESX host (keep_registered = true)")
		return
	}

	if remoteDriver, ok := driver.(RemoteDriver); ok {
		if s.SkipExport && !cancelled && !halted {
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
			start := time.Now()
			for {
				destroyed, err := remoteDriver.IsDestroyed()
				if destroyed {
					break
				}
				log.Printf("error destroying vm: %s", err)
				time.Sleep(1 * time.Second)
				if time.Since(start) >= time.Duration(30*time.Minute) {
					ui.Error("Error unregistering VM; timed out. You may " +
						"need to manually clean up your machine")
					break
				}
			}
		}
	}
}
