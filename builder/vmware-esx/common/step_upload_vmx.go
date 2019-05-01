package common

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step upload the VMX to the remote host
//
// Uses:
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepUploadVMX struct{}

func (c *StepUploadVMX) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	// Take the driver, convert it to a remote driver to upload the vmx
	remoteDriver, ok := driver.(RemoteDriver)
	if ok {
		remoteVmxPath := filepath.ToSlash(filepath.Join(fmt.Sprintf("%s", remoteDriver), filepath.Base(vmxPath)))
		if err := remoteDriver.upload(remoteVmxPath, vmxPath); err != nil {
			state.Put("error", fmt.Errorf("Error writing VMX: %s", err))
			return multistep.ActionHalt
		}

	} else {
		// This shouldn't ever happen, but its safer to not ignore the type assertion above.
		ui.Error(fmt.Sprintf("Driver does not implement the RemoteDriver interface."))
	}

	// Try and reload the VM now that it's been uploaded
	if err := remoteDriver.ReloadVM(); err != nil {
		ui.Error(fmt.Sprintf("Error reload VM: %s", err))
	}

	// Should be good to go
	return multistep.ActionContinue
}

func (StepUploadVMX) Cleanup(multistep.StateBag) {}
