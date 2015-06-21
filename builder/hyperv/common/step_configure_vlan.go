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
)


const(
	vlanId = "1724"
)

type StepConfigureVlan struct {
}

func (s *StepConfigureVlan) Run(state multistep.StateBag) multistep.StepAction {
	//config := state.Get("config").(*config)
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error configuring vlan: %s"
	vmName := state.Get("vmName").(string)
	switchName := state.Get("SwitchName").(string)

	ui.Say("Configuring vlan...")

	err := hyperv.SetNetworkAdapterVlanId(switchName, vlanId)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = hyperv.SetVirtualMachineVlanId(vmName, vlanId)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigureVlan) Cleanup(state multistep.StateBag) {
	//do nothing
}
