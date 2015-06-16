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

type StepDisableVlan struct {
}

func (s *StepDisableVlan) Run(state multistep.StateBag) multistep.StepAction {
	//	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error disabling vlan: %s"
	vmName := state.Get("vmName").(string)
	switchName := state.Get("SwitchName").(string)
	var err error

	ui.Say("Disabling vlan...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Set-VMNetworkAdapterVlan -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Untagged")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	blockBuffer.Reset()
	blockBuffer.WriteString("Set-VMNetworkAdapterVlan -ManagementOS -VMNetworkAdapterName '")
	blockBuffer.WriteString(switchName)
	blockBuffer.WriteString("' -Untagged")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepDisableVlan) Cleanup(state multistep.StateBag) {
	//do nothing
}
