package common

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step clones an existing virtual machine.
//
// Produces:
//   VMName string - The name of the VM
type StepCloneVM struct {
	CloneFromVMXCPath              string
	CloneFromVMName                string
	CloneFromSnapshotName          string
	CloneAllSnapshots              bool
	VMName                         string
	SwitchName                     string
	RamSize                        uint
	Cpu                            uint
	EnableMacSpoofing              bool
	EnableDynamicMemory            bool
	EnableSecureBoot               bool
	EnableVirtualizationExtensions bool
}

func (s *StepCloneVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Cloning virtual machine...")

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

	// convert the MB to bytes
	ramSize := int64(s.RamSize * 1024 * 1024)

	err := driver.CloneVirtualMachine(s.CloneFromVMXCPath, s.CloneFromVMName, s.CloneFromSnapshotName, s.CloneAllSnapshots, s.VMName, path, harddrivePath, ramSize, s.SwitchName)
	if err != nil {
		err := fmt.Errorf("Error cloning virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = driver.SetVirtualMachineCpuCount(s.VMName, s.Cpu)
	if err != nil {
		err := fmt.Errorf("Error creating setting virtual machine cpu: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if s.EnableDynamicMemory {
		err = driver.SetVirtualMachineDynamicMemory(s.VMName, s.EnableDynamicMemory)
		if err != nil {
			err := fmt.Errorf("Error creating setting virtual machine dynamic memory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if s.EnableMacSpoofing {
		err = driver.SetVirtualMachineMacSpoofing(s.VMName, s.EnableMacSpoofing)
		if err != nil {
			err := fmt.Errorf("Error creating setting virtual machine mac spoofing: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	generation, err := driver.GetVirtualMachineGeneration(s.VMName)
	if err != nil {
		err := fmt.Errorf("Error detecting vm generation: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if generation == 2 {
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
			err := fmt.Errorf("Error creating setting virtual machine virtualization extensions: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.VMName)

	return multistep.ActionContinue
}

func (s *StepCloneVM) Cleanup(state multistep.StateBag) {
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
