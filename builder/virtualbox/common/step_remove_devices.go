package common

import (
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
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
type StepRemoveDevices struct{}

func (s *StepRemoveDevices) Run(state multistep.StateBag) multistep.StepAction {
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

		var vboxErr error
		// Retry for 10 minutes to remove the floppy controller.
		log.Printf("Trying for 10 minutes to remove floppy controller.")
		err := common.Retry(15, 15, 40, func() (bool, error) {
			// Don't forget to remove the floppy controller as well
			command = []string{
				"storagectl", vmName,
				"--name", "Floppy Controller",
				"--remove",
			}
			vboxErr = driver.VBoxManage(command...)
			if vboxErr != nil {
				log.Printf("Error removing floppy controller. Retrying.")
				return false, nil
			}
			return true, nil
		})
		if err == common.RetryExhaustedError {
			err := fmt.Errorf("Error removing floppy controller: %s", vboxErr)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

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

	if _, ok := state.GetOk("guest_additions_attached"); ok {
		ui.Message("Removing guest additions drive...")
		command := []string{
			"storageattach", vmName,
			"--storagectl", "IDE Controller",
			"--port", "1",
			"--device", "0",
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
