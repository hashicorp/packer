// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepConfigureVlan struct {
	VlanId       string
	SwitchVlanId string
}

func (s *StepConfigureVlan) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error configuring vlan: %s"
	vmName := state.Get("vmName").(string)
	switchName := state.Get("SwitchName").(string)
	vlanId := s.VlanId
	switchVlanId := s.SwitchVlanId

	ui.Say("Configuring vlan...")

	if switchVlanId != "" {
		err := driver.SetNetworkAdapterVlanId(switchName, vlanId)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if vlanId != "" {
		err := driver.SetVirtualMachineVlanId(vmName, vlanId)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigureVlan) Cleanup(state multistep.StateBag) {
	//do nothing
}
