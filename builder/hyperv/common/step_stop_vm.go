// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepStopVm struct {
}

func (s *StepStopVm) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error stopping vm: %s"
	vmName := state.Get("vmName").(string)

	ui.Say("Stopping vm...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Stop-VM '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("'}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepStopVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
