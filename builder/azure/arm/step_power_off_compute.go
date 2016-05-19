// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/packer"
)

type StepPowerOffCompute struct {
	client   *AzureClient
	powerOff func(resourceGroupName string, computeName string, cancelCh <-chan struct{}) error
	say      func(message string)
	error    func(e error)
}

func NewStepPowerOffCompute(client *AzureClient, ui packer.Ui) *StepPowerOffCompute {
	var step = &StepPowerOffCompute{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.powerOff = step.powerOffCompute
	return step
}

func (s *StepPowerOffCompute) powerOffCompute(resourceGroupName string, computeName string, cancelCh <-chan struct{}) error {
	_, err := s.client.PowerOff(resourceGroupName, computeName, cancelCh)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepPowerOffCompute) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Powering off machine ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	result := common.StartInterruptibleTask(
		func() bool { return common.IsStateCancelled(state) },
		func(cancelCh <-chan struct{}) error { return s.powerOff(resourceGroupName, computeName, cancelCh) })

	return processInterruptibleResult(result, s.error, state)
}

func (*StepPowerOffCompute) Cleanup(multistep.StateBag) {
}
