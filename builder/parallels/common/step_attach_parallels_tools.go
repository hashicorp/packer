package common

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// StepAttachParallelsTools is a step that attaches Parallels Tools ISO image
// as an inserted CD onto the virtual machine.
//
// Uses:
//   driver Driver
//   parallels_tools_path string
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepAttachParallelsTools struct {
	cdromDevice        string
	ParallelsToolsMode string
}

// Run adds a virtual CD-ROM device to the VM and attaches Parallels Tools ISO image.
// If ISO image is not specified, then this step will be skipped.
func (s *StepAttachParallelsTools) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// If we're not attaching the guest additions then just return
	if s.ParallelsToolsMode != ParallelsToolsModeAttach {
		log.Println("Not attaching parallels tools since we're uploading.")
		return multistep.ActionContinue
	}

	// Get the Parallels Tools path on the host machine
	parallelsToolsPath := state.Get("parallels_tools_path").(string)

	// Attach the guest additions to the computer
	ui.Say("Attaching Parallels Tools ISO to the new CD/DVD drive...")

	cdrom, err := driver.DeviceAddCDROM(vmName, parallelsToolsPath)

	if err != nil {
		err = fmt.Errorf("Error attaching Parallels Tools ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Track the device name so that we can can delete later
	s.cdromDevice = cdrom

	return multistep.ActionContinue
}

// Cleanup removes the virtual CD-ROM device attached to the VM.
func (s *StepAttachParallelsTools) Cleanup(state multistep.StateBag) {
	if s.cdromDevice == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	log.Println("Detaching Parallels Tools ISO...")

	command := []string{
		"set", vmName,
		"--device-del", s.cdromDevice,
	}

	if err := driver.Prlctl(command...); err != nil {
		ui.Error(fmt.Sprintf("Error detaching Parallels Tools ISO: %s", err))
	}
}
