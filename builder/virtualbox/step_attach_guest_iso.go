package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step attaches the ISO to the virtual machine.
//
// Uses:
//
// Produces:
type stepAttachGuestISO struct {
	diskPath string
}

func (s *stepAttachGuestISO) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	guestAdditionsPath := state.Get("guest_additions_path").(string)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Attach the disk to the controller
	command := []string{
		"storageattach", vmName,
		"--storagectl", "IDE Controller",
		"--port", "1",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", guestAdditionsPath,
	}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error attaching ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from VirtualBox later
	s.diskPath = guestAdditionsPath

	return multistep.ActionContinue
}

func (s *stepAttachGuestISO) Cleanup(state multistep.StateBag) {
	if s.diskPath == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	command := []string{
		"storageattach", vmName,
		"--storagectl", "IDE Controller",
		"--port", "1",
		"--device", "0",
		"--medium", "none",
	}

	if err := driver.VBoxManage(command...); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering ISO: %s", err))
	}
}
