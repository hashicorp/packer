// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"

	"github.com/mitchellh/packer/builder/azure/common/constants"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepGetOSDisk struct {
	client *AzureClient
	query  func(resourceGroupName string, computeName string) (compute.VirtualMachine, error)
	say    func(message string)
	error  func(e error)
}

func NewStepGetOSDisk(client *AzureClient, ui packer.Ui) *StepGetOSDisk {
	var step = &StepGetOSDisk{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.query = step.queryCompute
	return step
}

func (s *StepGetOSDisk) queryCompute(resourceGroupName string, computeName string) (compute.VirtualMachine, error) {
	return s.client.VirtualMachinesClient.Get(resourceGroupName, computeName, "")
}

func (s *StepGetOSDisk) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Querying the machine's properties ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	vm, err := s.query(resourceGroupName, computeName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	s.say(fmt.Sprintf(" -> OS Disk           : '%s'", *vm.Properties.StorageProfile.OsDisk.Vhd.URI))
	state.Put(constants.ArmOSDiskVhd, *vm.Properties.StorageProfile.OsDisk.Vhd.URI)

	return multistep.ActionContinue
}

func (*StepGetOSDisk) Cleanup(multistep.StateBag) {
}
