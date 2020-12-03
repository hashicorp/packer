package common

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step upload the VMX to the remote host
//
// Uses:
//   driver Driver
//   ui     packersdk.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepUploadVMX struct {
	RemoteType string
}

func (c *StepUploadVMX) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	ui := state.Get("ui").(packersdk.Ui)
	vmxPath := state.Get("vmx_path").(string)

	if c.RemoteType == "esx5" {
		remoteDriver, ok := driver.(RemoteDriver)
		if ok {
			remoteVmxPath := filepath.ToSlash(filepath.Join(fmt.Sprintf("%s", remoteDriver), filepath.Base(vmxPath)))
			if err := remoteDriver.upload(remoteVmxPath, vmxPath, nil); err != nil {
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
