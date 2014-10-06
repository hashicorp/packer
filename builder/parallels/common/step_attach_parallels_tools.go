package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step attaches the Parallels Tools as an inserted CD onto
// the virtual machine.
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

func (s *StepAttachParallelsTools) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// If we're not attaching the guest additions then just return
	if s.ParallelsToolsMode != ParallelsToolsModeAttach {
		log.Println("Not attaching parallels tools since we're uploading.")
		return multistep.ActionContinue
	}

	// Get the Paralells Tools path on the host machine
	parallelsToolsPath := state.Get("parallels_tools_path").(string)

	// Attach the guest additions to the computer
	ui.Say("Attaching Parallels Tools ISO to the new CD/DVD drive...")

	cdrom, err := driver.DeviceAddCdRom(vmName, parallelsToolsPath)

	if err != nil {
		err := fmt.Errorf("Error attaching Parallels Tools ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Track the device name so that we can can delete later
	s.cdromDevice = cdrom

	return multistep.ActionContinue
}

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
