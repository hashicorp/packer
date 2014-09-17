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
//   driver Driver
//   isoPath string
//   ui packer.Ui
//   vmName string
//
// Produces:
type stepAttachISO struct {
	cdromDevice string
}

func (s *stepAttachISO) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(parallelscommon.Driver)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Attach the disk to the controller
	ui.Say("Attaching ISO to the new CD/DVD drive...")

	cdrom, err := driver.DeviceAddCdRom(vmName, isoPath)

	if err != nil {
		err := fmt.Errorf("Error attaching ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Track the device name so that we can can delete later
	s.cdromDevice = cdrom

	return multistep.ActionContinue
}

func (s *stepAttachISO) Cleanup(state multistep.StateBag) {
	if s.cdromDevice == "" {
		return
	}

	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	log.Println("Detaching ISO...")

	command := []string{
		"set", vmName,
		"--device-del", s.cdromDevice,
	}

	if err := driver.Prlctl(command...); err != nil {
		ui.Error(fmt.Sprintf("Error detaching ISO: %s", err))
	}
}
