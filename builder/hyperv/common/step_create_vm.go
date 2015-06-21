// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/powershell/hyperv"
	"strconv"
)

// This step creates the actual virtual machine.
//
// Produces:
//   VMName string - The name of the VM
type StepCreateVM struct {
	VMName     string
	SwitchName string
	RamSizeMB  uint
	DiskSize   uint
	Generation uint
	Cpu        uint
}

func (s *StepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating virtual machine...")

	path := state.Get("packerTempDir").(string)

	// convert the MB to bytes
	ramBytes := int64(s.RamSizeMB * 1024 * 1024)
	diskSizeBytes := int64(s.DiskSize * 1024 * 1024)

	ram := strconv.FormatInt(ramBytes, 10)
	diskSize := strconv.FormatInt(diskSizeBytes, 10)
	switchName := s.SwitchName
	generation := strconv.FormatInt(int64(s.Generation), 10)
	cpu := strconv.FormatInt(int64(s.Cpu), 10)

	err := hyperv.CreateVirtualMachine(s.VMName, path, ram, diskSize, switchName, generation)
	if err != nil {
		err := fmt.Errorf("Error creating virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = hyperv.SetVirtualMachineCpu(s.VMName, cpu)
	if err != nil {
		err := fmt.Errorf("Error creating setting virtual machine cpu: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.VMName)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	if s.VMName == "" {
		return
	}

	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Unregistering and deleting virtual machine...")

	err := hyperv.DeleteVirtualMachine(s.VMName)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
