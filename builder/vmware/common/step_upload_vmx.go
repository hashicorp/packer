package common

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
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
type StepUploadVMX struct {
	RemoteType string
}

func (c *StepUploadVMX) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	if c.RemoteType == "esx5" {
		remoteDriver, ok := driver.(RemoteDriver)
		if ok {
			remoteVmxPath := filepath.ToSlash(filepath.Join(fmt.Sprintf("%s", remoteDriver), filepath.Base(vmxPath)))
			log.Printf("Uploading VMX file from %s to %s", vmxPath, remoteVmxPath)
			if err := remoteDriver.Upload(remoteVmxPath, vmxPath); err != nil {
				state.Put("error", fmt.Errorf("Error writing VMX: %s", err))
				return multistep.ActionHalt
			}
		}
		if err := remoteDriver.ReloadVM(); err != nil {
			ui.Error(fmt.Sprintf("Error reload VM: %s", err))
		}
	}

	return multistep.ActionContinue
}

func (StepUploadVMX) Cleanup(multistep.StateBag) {}
