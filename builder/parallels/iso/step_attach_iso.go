package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	parallelscommon "github.com/mitchellh/packer/builder/parallels/common"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step attaches the ISO to the virtual machine.
//
// Uses:
//
// Produces:
type stepAttachISO struct {
	diskPath string
}

func (s *stepAttachISO) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(parallelscommon.Driver)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Attach the disk to the controller
	ui.Say("Attaching ISO onto IDE controller...")
	command := []string{
		"set", vmName,
		"--device-set", "cdrom0",
		"--image", isoPath,
		"--enable", "--connect",
	}
	if err := driver.Prlctl(command...); err != nil {
		err := fmt.Errorf("Error attaching ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set some state so we know to remove
	state.Put("attachedIso", true)

	return multistep.ActionContinue
}

func (s *stepAttachISO) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("attachedIso"); !ok {
		return
	}

	driver := state.Get("driver").(parallelscommon.Driver)
	vmName := state.Get("vmName").(string)

	command := []string{
		"set", vmName,
		"--device-set", "cdrom0",
		"--enable", "--disconnect",
	}

	// Remove the ISO, ignore errors
	log.Println("Detaching ISO...")
	driver.Prlctl(command...)
}
