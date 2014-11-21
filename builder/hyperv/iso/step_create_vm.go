// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	hypervcommon "github.com/mitchellh/packer/builder/hyperv/common"
	"github.com/mitchellh/packer/packer"
)

// This step creates the actual virtual machine.
//
// Produces:
//   vmName string - The name of the VM
type StepCreateVM struct {
	vmName string
}

func (s *StepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*iso_config)
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating virtual machine...")

	vmName := config.VMName
//	path :=	config.OutputDir
	path :=	state.Get("packerTempDir").(string)

	ram := config.RamSizeMB
	diskSize := config.DiskSizeGB
	switchName := config.SwitchName

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {New-VM -Name '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Path '")
	blockBuffer.WriteString(path)
	blockBuffer.WriteString("' -MemoryStartupBytes ")
	blockBuffer.WriteString(fmt.Sprintf("%vMB",ram))
	blockBuffer.WriteString(" -NewVHDPath '")
	blockBuffer.WriteString(path)
	blockBuffer.WriteString("/")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString(".vhdx'")
	blockBuffer.WriteString(" -NewVHDSizeBytes ")
	blockBuffer.WriteString(fmt.Sprintf("%vGB",diskSize))
	blockBuffer.WriteString(" -SwitchName '")
	blockBuffer.WriteString(switchName)
	blockBuffer.WriteString("'}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf("Error creating virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the VM name property on the first command
	if s.vmName == "" {
		s.vmName = vmName
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.vmName)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering and deleting virtual machine...")

	var err error = nil

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Remove-VM –Name '")
	blockBuffer.WriteString(s.vmName)
	blockBuffer.WriteString("' -Force }")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
