package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step attaches the ISO to the virtual machine.
//
// Uses:
//
// Produces:
type stepAttachISO struct{
	diskPath string
}

func (s *stepAttachISO) Run(state map[string]interface{}) multistep.StepAction {
	driver := state["driver"].(Driver)
	isoPath := state["iso_path"].(string)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

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
		ui.Error(fmt.Sprintf("Error attaching hard drive: %s", err))
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from VirtualBox later
	s.diskPath = isoPath

	time.Sleep(15 * time.Second)
	return multistep.ActionContinue
}

func (s *stepAttachISO) Cleanup(state map[string]interface{}) {
	if s.diskPath == "" {
		return
	}

	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	if err := driver.VBoxManage("closemedium", "disk", s.diskPath); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering ISO: %s", err))
	}
}
