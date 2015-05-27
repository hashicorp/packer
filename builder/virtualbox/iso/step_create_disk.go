package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"strconv"
	"strings"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	format := "VDI"
	path := filepath.Join(config.OutputDir, fmt.Sprintf("%s.%s", config.VMName, strings.ToLower(format)))

	command := []string{
		"createhd",
		"--filename", path,
		"--size", strconv.FormatUint(uint64(config.DiskSize), 10),
		"--format", format,
		"--variant", "Standard",
	}

	ui.Say("Creating hard drive...")
	err := driver.VBoxManage(command...)
	if err != nil {
		err := fmt.Errorf("Error creating hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Add the IDE controller so we can later attach the disk.
	// When the hard disk controller is not IDE, this device is still used
	// by VirtualBox to deliver the guest extensions.
	err = driver.VBoxManage("storagectl", vmName, "--name", "IDE Controller", "--add", "ide")
	if err != nil {
		err := fmt.Errorf("Error creating disk controller: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Add a SATA controller if we were asked to use SATA. We still attach
	// the IDE controller above because some other things (disks) require
	// that.
	if config.HardDriveInterface == "sata" || config.ISOInterface == "sata" {
		if err := driver.CreateSATAController(vmName, "SATA Controller"); err != nil {
			err := fmt.Errorf("Error creating disk controller: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if config.HardDriveInterface == "scsi" {
		if err := driver.CreateSCSIController(vmName, "SCSI Controller"); err != nil {
			err := fmt.Errorf("Error creating disk controller: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Attach the disk to the controller
	controllerName := "IDE Controller"
	if config.HardDriveInterface == "sata" {
		controllerName = "SATA Controller"
	}

	if config.HardDriveInterface == "scsi" {
		controllerName = "SCSI Controller"
	}

	command = []string{
		"storageattach", vmName,
		"--storagectl", controllerName,
		"--port", "0",
		"--device", "0",
		"--type", "hdd",
		"--medium", path,
	}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error attaching hard drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
