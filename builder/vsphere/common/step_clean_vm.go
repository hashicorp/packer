package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// This step cleans up the VM by removing or changing this prior to
// being ready for use.
//
// Uses:
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepCleanVM struct {
	CustomData map[string]string
}

func (s StepCleanVM) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Cleaning VM prior to finishing up...")

	// Delete the floppy0 entries so the floppy is no longer mounted
	if floppyDevice, ok := state.GetOk("floppy_device"); ok {
		ui.Message("Unmounting floppy from VMX...")
		if err := driver.RemoveFloppy(floppyDevice.(string)); err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error removing floppy: %s", err))
			return multistep.ActionHalt
		}
	}

	// Set custom data
	for k, v := range s.CustomData {
		log.Printf("Setting VMX: '%s' = '%s'", k, v)
		k = strings.ToLower(k)
		if err := driver.VMChange(fmt.Sprintf("%s=%s", k, v)); err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error changing VM: %s", err))
			return multistep.ActionHalt
		}
	}

	//Remove Cdrom if necessary
	if cdromDevice, ok := state.GetOk("cdrom_device"); ok {
		ui.Message("Detaching ISO from CD-ROM device...")
		if err := driver.UnmountISO(cdromDevice.(string)); err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error detaching ISO from VM: %s", err))
			return multistep.ActionHalt
		}
	}

	ui.Message("Disabling VNC server...")
	if err := driver.VNCDisable(); err != nil {
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error disabling VNC server: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (StepCleanVM) Cleanup(multistep.StateBag) {}
