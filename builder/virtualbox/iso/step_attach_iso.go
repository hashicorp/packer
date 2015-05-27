package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
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
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vboxcommon.Driver)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	controllerName := "IDE Controller"
	port := "0"
	device := "1"
	if config.ISOInterface == "sata" {
		controllerName = "SATA Controller"
		port = "1"
		device = "0"
	}

	// Attach the disk to the controller
	command := []string{
		"storageattach", vmName,
		"--storagectl", controllerName,
		"--port", port,
		"--device", device,
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

	// Set some state so we know to remove
	state.Put("attachedIso", true)
	if controllerName == "SATA Controller" {
		state.Put("attachedIsoOnSata", true)
	}

	return multistep.ActionContinue
}

func (s *stepAttachISO) Cleanup(state multistep.StateBag) {
	if s.diskPath == "" {
		return
	}

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vboxcommon.Driver)
	vmName := state.Get("vmName").(string)

	controllerName := "IDE Controller"
	port := "0"
	device := "1"
	if config.ISOInterface == "sata" {
		controllerName = "SATA Controller"
		port = "1"
		device = "0"
	}

	command := []string{
		"storageattach", vmName,
		"--storagectl", controllerName,
		"--port", port,
		"--device", device,
		"--medium", "none",
	}

	// Remove the ISO. Note that this will probably fail since
	// stepRemoveDevices does this as well. No big deal.
	driver.VBoxManage(command...)
}
