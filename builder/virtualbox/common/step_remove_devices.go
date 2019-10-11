package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step removes any devices (floppy disks, ISOs, etc.) from the
// machine that we may have added.
//
// Uses:
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepRemoveDevices struct {
	Bundling                VBoxBundleConfig
	GuestAdditionsInterface string
}

func (s *StepRemoveDevices) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Remove the attached floppy disk, if it exists
	if _, ok := state.GetOk("floppy_path"); ok {
		ui.Message("Removing floppy drive...")
		command := []string{
			"storageattach", vmName,
			"--storagectl", "Floppy Controller",
			"--port", "0",
			"--device", "0",
			"--medium", "none",
		}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error removing floppy: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Retry for 10 minutes to remove the floppy controller.
		log.Printf("Trying for 10 minutes to remove floppy controller.")
		err := retry.Config{
			Tries:      40,
			RetryDelay: (&retry.Backoff{InitialBackoff: 15 * time.Second, MaxBackoff: 15 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			// Don't forget to remove the floppy controller as well
			command = []string{
				"storagectl", vmName,
				"--name", "Floppy Controller",
				"--remove",
			}
			err := driver.VBoxManage(command...)
			if err != nil {
				log.Printf("Error removing floppy controller. Retrying.")
			}
			return err
		})
		if err != nil {
			err := fmt.Errorf("Error removing floppy controller: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if !s.Bundling.BundleISO {
		if _, ok := state.GetOk("attachedIso"); ok {
			controllerName := "IDE Controller"
			port := "0"
			device := "1"
			if _, ok := state.GetOk("attachedIsoOnSata"); ok {
				controllerName = "SATA Controller"
				port = "1"
				device = "0"
			}

			command := []string{
				"storageattach", vmName,
				"--storagectl", controllerName,
				"--port", port,
				"--device", device,
				"--medium", "none",
			}

			if err := driver.VBoxManage(command...); err != nil {
				err := fmt.Errorf("Error detaching ISO: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	if _, ok := state.GetOk("guest_additions_attached"); ok {
		ui.Message("Removing guest additions drive...")
		controllerName := "IDE Controller"
		port := "1"
		device := "0"
		if s.GuestAdditionsInterface == "sata" {
			controllerName = "SATA Controller"
			port = "2"
			device = "0"
		}
		command := []string{
			"storageattach", vmName,
			"--storagectl", controllerName,
			"--port", port,
			"--device", device,
			"--medium", "none",
		}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error removing guest additions: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepRemoveDevices) Cleanup(state multistep.StateBag) {
}
