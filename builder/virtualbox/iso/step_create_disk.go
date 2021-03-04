package iso

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"

	"path/filepath"
	"strconv"
	"strings"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
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
		if err := driver.CreateSATAController(vmName, "SATA Controller", config.SATAPortCount); err != nil {
			err := fmt.Errorf("Error creating disk controller: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Add a VirtIO controller if we were asked to use VirtIO. We still attach
	// the VirtIO controller above because some other things (disks) require
	// that.
	if config.HardDriveInterface == "virtio" || config.ISOInterface == "virtio" {
		if err := driver.CreateVirtIOController(vmName, "VirtIO Controller"); err != nil {
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
	} else if config.HardDriveInterface == "pcie" {
		if err := driver.CreateNVMeController(vmName, "NVMe Controller", config.NVMePortCount); err != nil {
			err := fmt.Errorf("Error creating NVMe controller: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Attach the disk to the controller
	controllerName := "IDE Controller"
	if config.HardDriveInterface == "sata" {
		controllerName = "SATA Controller"
	} else if config.HardDriveInterface == "scsi" {
		controllerName = "SCSI Controller"
	} else if config.HardDriveInterface == "virtio" {
		controllerName = "VirtIO Controller"
	} else if config.HardDriveInterface == "pcie" {
		controllerName = "NVMe Controller"
	}

	nonrotational := "off"
	if config.HardDriveNonrotational {
		nonrotational = "on"
	}

	discard := "off"
	if config.HardDriveDiscard {
		discard = "on"
	}

	command = []string{
		"storageattach", vmName,
		"--storagectl", controllerName,
		"--port", "0",
		"--device", "0",
		"--type", "hdd",
		"--medium", path,
		"--nonrotational", nonrotational,
		"--discard", discard,
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
