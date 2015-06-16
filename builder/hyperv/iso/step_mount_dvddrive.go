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


type StepMountDvdDrive struct {
	path string
}

func (s *StepMountDvdDrive) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*iso_config)
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error mounting dvd drive: %s"
	vmName := state.Get("vmName").(string)
	isoPath := config.RawSingleISOUrl

	ui.Say("Mounting dvd drive...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Set-VMDvdDrive -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Path '")
	blockBuffer.WriteString(isoPath)
	blockBuffer.WriteString("'}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.path = isoPath

	return multistep.ActionContinue
}

func (s *StepMountDvdDrive) Cleanup(state multistep.StateBag) {
	if s.path == "" {
		return
	}

	errorMsg := "Error unmounting dvd drive: %s"

	vmName := state.Get("vmName").(string)
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unmounting dvd drive...")

	var err error = nil

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Set-VMDvdDrive -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Path $null}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf(errorMsg, err))
	}
}
