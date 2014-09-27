package common

import (
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
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
//   attachedToolsIso boolean
type StepAttachParallelsTools struct {
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
	ui.Say("Attaching Parallels Tools ISO onto IDE controller...")
	command := []string{
		"set", vmName,
		"--device-add", "cdrom",
		"--image", parallelsToolsPath,
	}
	if err := driver.Prlctl(command...); err != nil {
		err := fmt.Errorf("Error attaching Parallels Tools: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set some state so we know to remove it later
	state.Put("attachedToolsIso", true)

	return multistep.ActionContinue
}

func (s *StepAttachParallelsTools) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("attachedToolsIso"); !ok {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if _, ok := state.GetOk("attachedIso"); ok {
		log.Println("Detaching Parallels Tools ISO...")
		command := []string{
			"set", vmName,
			"--device-del", "cdrom1",
		}

		if err := driver.Prlctl(command...); err != nil {
			ui.Error(fmt.Sprintf("Error detaching Parallels Tools ISO: %s", err))
		}
	}
}
