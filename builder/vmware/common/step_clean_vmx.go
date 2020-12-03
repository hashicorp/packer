package common

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step cleans up the VMX by removing or changing this prior to
// being ready for use.
//
// Uses:
//   ui     packersdk.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepCleanVMX struct {
	RemoveEthernetInterfaces bool
	VNCEnabled               bool
}

func (s StepCleanVMX) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vmxPath := state.Get("vmx_path").(string)

	ui.Say("Cleaning VMX prior to finishing up...")

	vmxData, err := ReadVMX(vmxPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading VMX: %s", err))
		return multistep.ActionHalt
	}

	// Grab our list of devices added during the build out of the statebag
	for _, device := range state.Get("temporaryDevices").([]string) {
		// Instead of doing this in one pass which would be more efficient,
		// we do it per device-type so that the logic appears to be the same
		// as the prior implementation.

		// Walk through all the devices that were temporarily added and figure
		// out which type it is in order to figure out how to disable it.
		// Right now only floppy, cdrom devices, ethernet, and devices that use
		// ".present" are supported.
		if strings.HasPrefix(device, "floppy") {
			// We can identify a floppy device because it begins with "floppy"
			ui.Message(fmt.Sprintf("Unmounting %s from VMX...", device))

			// Delete the floppy%d entries so the floppy is no longer mounted
			for k := range vmxData {
				if strings.HasPrefix(k, fmt.Sprintf("%s.", device)) {
					log.Printf("Deleting key for floppy device: %s", k)
					delete(vmxData, k)
				}
			}
			vmxData[fmt.Sprintf("%s.present", device)] = "FALSE"

		} else if strings.HasPrefix(vmxData[fmt.Sprintf("%s.devicetype", device)], "cdrom-") {
			// We can identify something is a cdrom if it has a ".devicetype"
			// attribute that begins with "cdrom-"
			ui.Message(fmt.Sprintf("Detaching ISO from CD-ROM device %s...", device))

			// Simply turn the CDROM device into a native cdrom instead of an iso
			vmxData[fmt.Sprintf("%s.devicetype", device)] = "cdrom-raw"
			vmxData[fmt.Sprintf("%s.filename", device)] = "auto detect"
			vmxData[fmt.Sprintf("%s.clientdevice", device)] = "TRUE"

		} else if strings.HasPrefix(device, "ethernet") && s.RemoveEthernetInterfaces {
			// We can identify an ethernet device because it begins with "ethernet"
			// Although we're supporting this, as of now it's not in use due
			// to these interfaces not ever being added to the "temporaryDevices" statebag.
			ui.Message(fmt.Sprintf("Removing %s interface...", device))

			// Delete the ethernet%d entries so the ethernet interface is removed.
			// This corresponds to the same logic defined below.
			for k := range vmxData {
				if strings.HasPrefix(k, fmt.Sprintf("%s.", device)) {
					log.Printf("Deleting key for ethernet device: %s", k)
					delete(vmxData, k)
				}
			}

		} else {

			// First check to see if we can simply disable the device
			if _, ok := vmxData[fmt.Sprintf("%s.present", device)]; ok {
				ui.Message(fmt.Sprintf("Disabling device %s of an unknown device type...", device))
				vmxData[fmt.Sprintf("%s.present", device)] = "FALSE"
			} else {
				// Okay, so this wasn't so simple. Let's just log info about the
				// device and not tamper with any of its keys
				log.Printf("Refusing to remove device due to being of an unsupported type: %s\n", device)
				for k := range vmxData {
					if strings.HasPrefix(k, fmt.Sprintf("%s.", device)) {
						log.Printf("Leaving unsupported device key: %s\n", k)
					}
				}
			}
		}
	}

	// Disable the VNC server if necessary
	if s.VNCEnabled {
		ui.Message("Disabling VNC server...")
		vmxData["remotedisplay.vnc.enabled"] = "FALSE"
	}

	// Disable any ethernet devices if necessary
	if s.RemoveEthernetInterfaces {
		ui.Message("Removing Ethernet Interfaces...")
		for k := range vmxData {
			if strings.HasPrefix(k, "ethernet") {
				log.Printf("Deleting key for ethernet device: %s", k)
				delete(vmxData, k)
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
