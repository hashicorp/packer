package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step attaches a floppy to the virtual machine.
//
// Uses:
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepAttachFloppy struct {
	floppyPath string
}

func (s *StepAttachFloppy) Run(state multistep.StateBag) multistep.StepAction {
	// Determine if we even have a floppy disk to attach
	var floppyPath string
	if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
		floppyPath = floppyPathRaw.(string)
	} else {
		log.Println("No floppy disk, not attaching.")
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Attaching floppy disk...")

	// Create the floppy disk controller
	command := []string{
		"set", vmName,
		"--device-add", "fdd",
		"--image", floppyPath,
	}
	if err := driver.Prlctl(command...); err != nil {
		state.Put("error", fmt.Errorf("Error adding floppy: %s", err))
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from Parallels later
	s.floppyPath = floppyPath

	return multistep.ActionContinue
}

func (s *StepAttachFloppy) Cleanup(state multistep.StateBag) {}
