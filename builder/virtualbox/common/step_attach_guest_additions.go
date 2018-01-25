package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step attaches the VirtualBox guest additions as a inserted CD onto
// the virtual machine.
//
// Uses:
//   config *config
//   driver Driver
//   guest_additions_path string
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepAttachGuestAdditions struct {
	attachedPath       string
	GuestAdditionsMode string
}

func (s *StepAttachGuestAdditions) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// If we're not attaching the guest additions then just return
	if s.GuestAdditionsMode != GuestAdditionsModeAttach {
		log.Println("Not attaching guest additions since we're uploading.")
		return multistep.ActionContinue
	}

	// Get the guest additions path since we're doing it
	guestAdditionsPath := state.Get("guest_additions_path").(string)

	// Attach the guest additions to the computer
	log.Println("Attaching guest additions ISO onto IDE controller...")
	command := []string{
		"storageattach", vmName,
		"--storagectl", "IDE Controller",
		"--port", "1",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", guestAdditionsPath,
	}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error attaching guest additions: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from VirtualBox later
	s.attachedPath = guestAdditionsPath
	state.Put("guest_additions_attached", true)

	return multistep.ActionContinue
}

func (s *StepAttachGuestAdditions) Cleanup(state multistep.StateBag) {
	if s.attachedPath == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)

	command := []string{
		"storageattach", vmName,
		"--storagectl", "IDE Controller",
		"--port", "1",
		"--device", "0",
		"--medium", "none",
	}

	// Remove the ISO. Note that this will probably fail since
	// stepRemoveDevices does this as well. No big deal.
	driver.VBoxManage(command...)
}
