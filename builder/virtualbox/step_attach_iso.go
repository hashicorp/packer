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
type stepAttachISO struct {
	diskPath string
}

func (s *stepAttachISO) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Attach the disk to the controller
	command := []string{
		"storageattach", vmName,
		"--storagectl", "IDE Controller",
		"--port", "0",
		"--device", "1",
		"--type", "dvddrive",
		"--medium", isoPath,
	}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error attaching ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from VirtualBox later
	s.diskPath = isoPath

	return multistep.ActionContinue
}

func (s *stepAttachISO) Cleanup(state multistep.StateBag) {
	if s.diskPath == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	command := []string{
		"storageattach", vmName,
		"--storagectl", "IDE Controller",
		"--port", "0",
		"--device", "1",
		"--medium", "none",
	}

	if err := driver.VBoxManage(command...); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering ISO: %s", err))
	}
}
