// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"fmt"
	"bytes"
	"os"
	"github.com/mitchellh/multistep"
	hypervcommon "github.com/mitchellh/packer/builder/hyperv/common"
	"github.com/mitchellh/packer/packer"
)


type StepUnmountFloppydrive struct {
	fileName string
	dir string
}

func (s *StepUnmountFloppydrive) Run(state multistep.StateBag) multistep.StepAction {
//	config := state.Get("config").(*config)
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Unmounting floppy drive: %s"
	vmName := state.Get("vmName").(string)
	packerTempDir :=  state.Get("packerTempDir").(string)


	ui.Say("Unmounting floppy drive...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Set-VMFloppyDiskDrive -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Path $null}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
//		return multistep.ActionHalt
	}


	floppyfile := packerTempDir + "/" + FloppyFileName
	err = os.Remove(floppyfile)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		//		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUnmountFloppydrive) Cleanup(state multistep.StateBag) {
	// do nothing
}
