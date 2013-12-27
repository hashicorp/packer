package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"regexp"
	"strings"
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

	if _, ok := state.GetOk("floppy_path"); ok {
		// Delete the floppy0 entries so the floppy is no longer mounted
		ui.Message("Unmounting floppy from VMX...")
		for k, _ := range vmxData {
			if strings.HasPrefix(k, "floppy0.") {
				log.Printf("Deleting key: %s", k)
				delete(vmxData, k)
			}
		}
		vmxData["floppy0.present"] = "FALSE"
	}

	if isoPathRaw, ok := state.GetOk("iso_path"); ok {
		isoPath := isoPathRaw.(string)

		ui.Message("Detaching ISO from CD-ROM device...")
		devRe := regexp.MustCompile(`^ide\d:\d\.`)
		for k, _ := range vmxData {
			match := devRe.FindString(k)
			if match == "" {
				continue
			}

			filenameKey := match + "filename"
			if filename, ok := vmxData[filenameKey]; ok {
				if filename == isoPath {
					// Change the CD-ROM device back to auto-detect to eject
					vmxData[filenameKey] = "auto detect"
					vmxData[match+"devicetype"] = "cdrom-raw"
				}
			}
		}
	}

	// Rewrite the VMX
	if err := WriteVMX(vmxPath, vmxData); err != nil {
		state.Put("error", fmt.Errorf("Error writing VMX: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (StepCleanVMX) Cleanup(multistep.StateBag) {}
