package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/retry"
)

// This step removes any devices (floppy disks, ISOs, etc.) from the
// machine that we may have added.
//
// Uses:
//   driver Driver
//   ui packersdk.Ui
//   vmName string
//
// Produces:
type StepRemoveDevices struct {
	Bundling VBoxBundleConfig
}

func (s *StepRemoveDevices) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
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

	var isoUnmountCommands map[string][]string
	isoUnmountCommandsRaw, ok := state.GetOk("disk_unmount_commands")
	if !ok {
		// No disks to unmount
		return multistep.ActionContinue
	} else {
		isoUnmountCommands = isoUnmountCommandsRaw.(map[string][]string)
	}

	for diskCategory, unmountCommand := range isoUnmountCommands {
		if diskCategory == "boot_iso" && s.Bundling.BundleISO {
			// skip the unmount if user wants to bundle the iso
			continue
		}

		if err := driver.VBoxManage(unmountCommand...); err != nil {
			err := fmt.Errorf("Error detaching ISO: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// log that we removed the isos, so we don't waste time trying to do it
	// in the step_attach_isos cleanup.
	state.Put("detached_isos", true)

	return multistep.ActionContinue
}

func (s *StepRemoveDevices) Cleanup(state multistep.StateBag) {
}
