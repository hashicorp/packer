package common

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step cleans up the VMX by removing or changing this prior to
// being ready for use.
//
// Uses:
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepCleanVMX struct{}

func (s StepCleanVMX) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	ui.Say("Cleaning VMX prior to finishing up...")

	vmxData, err := ReadVMX(vmxPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading VMX: %s", err))
		return multistep.ActionHalt
	}

	// Delete the floppy0 entries so the floppy is no longer mounted
	ui.Message("Unmounting floppy from VMX...")
	for k, _ := range vmxData {
		if strings.HasPrefix(k, "floppy0.") {
			log.Printf("Deleting key: %s", k)
			delete(vmxData, k)
		}
	}
	vmxData["floppy0.present"] = "FALSE"

	devRe := regexp.MustCompile(`^ide\d:\d\.`)
	for k, v := range vmxData {
		ide := devRe.FindString(k)
		if ide == "" || v != "cdrom-image" {
			continue
		}

		ui.Message("Detaching ISO from CD-ROM device...")

		vmxData[ide+"devicetype"] = "cdrom-raw"
		vmxData[ide+"filename"] = "auto detect"
		vmxData[ide+"clientdevice"] = "TRUE"
	}

	ui.Message("Disabling VNC server...")
	vmxData["remotedisplay.vnc.enabled"] = "FALSE"

	// Rewrite the VMX
	if err := WriteVMX(vmxPath, vmxData); err != nil {
		state.Put("error", fmt.Errorf("Error writing VMX: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (StepCleanVMX) Cleanup(multistep.StateBag) {}
