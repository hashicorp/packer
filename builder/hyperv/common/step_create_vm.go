package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step creates the actual virtual machine.
//
// Produces:
//   VMName string - The name of the VM
type StepCreateVM struct {
	VMName                         string
	SwitchName                     string
	HarddrivePath                  string
	RamSize                        uint
	DiskSize                       uint
	Generation                     uint
	Cpu                            uint
	EnableMacSpoofing              bool
	EnableDynamicMemory            bool
	EnableSecureBoot               bool
	EnableVirtualizationExtensions bool
	AdditionalDiskSize             []uint
	DifferencingDisk               bool
}

func (s *StepCreateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating virtual machine...")

	path := state.Get("packerTempDir").(string)

	// Determine if we even have an existing virtual harddrive to attach
	harddrivePath := ""
	if harddrivePathRaw, ok := state.GetOk("iso_path"); ok {
		extension := strings.ToLower(filepath.Ext(harddrivePathRaw.(string)))
		if extension == ".vhd" || extension == ".vhdx" {
			harddrivePath = harddrivePathRaw.(string)
		} else {
			log.Println("No existing virtual harddrive, not attaching.")
		}
	} else {
		log.Println("No existing virtual harddrive, not attaching.")
	}

	vhdPath := state.Get("packerVhdTempDir").(string)
	// convert the MB to bytes
	ramSize := int64(s.RamSize * 1024 * 1024)
	diskSize := int64(s.DiskSize * 1024 * 1024)

	err := driver.CreateVirtualMachine(s.VMName, path, harddrivePath, vhdPath, ramSize, diskSize, s.SwitchName, s.Generation, s.DifferencingDisk)
	if err != nil {
		err := fmt.Errorf("Error creating virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = driver.SetVirtualMachineCpuCount(s.VMName, s.Cpu)
	if err != nil {
		err := fmt.Errorf("Error setting virtual machine cpu count: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = driver.SetVirtualMachineDynamicMemory(s.VMName, s.EnableDynamicMemory)
	if err != nil {
		err := fmt.Errorf("Error setting virtual machine dynamic memory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if s.EnableMacSpoofing {
		err = driver.SetVirtualMachineMacSpoofing(s.VMName, s.EnableMacSpoofing)
		if err != nil {
			err := fmt.Errorf("Error setting virtual machine mac spoofing: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if s.Generation == 2 {
		err = driver.SetVirtualMachineSecureBoot(s.VMName, s.EnableSecureBoot)
		if err != nil {
			err := fmt.Errorf("Error setting secure boot: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if s.EnableVirtualizationExtensions {
		//This is only supported on Windows 10 and Windows Server 2016 onwards
		err = driver.SetVirtualMachineVirtualizationExtensions(s.VMName, s.EnableVirtualizationExtensions)
		if err != nil {
			err := fmt.Errorf("Error setting virtual machine virtualization extensions: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if len(s.AdditionalDiskSize) > 0 {
		for index, size := range s.AdditionalDiskSize {
			diskSize := int64(size * 1024 * 1024)
			diskFile := fmt.Sprintf("%s-%d.vhdx", s.VMName, index)
			err = driver.AddVirtualMachineHardDrive(s.VMName, vhdPath, diskFile, diskSize, "SCSI")
			if err != nil {
				err := fmt.Errorf("Error creating and attaching additional disk drive: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.VMName)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	if s.VMName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Unregistering and deleting virtual machine...")

	err := driver.DeleteVirtualMachine(s.VMName)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
