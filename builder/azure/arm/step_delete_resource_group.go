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

type StepDeleteResourceGroup struct {
	client *AzureClient
	delete func(resourceGroupName string, cancelCh <-chan struct{}) error
	say    func(message string)
	error  func(e error)
}

func NewStepDeleteResourceGroup(client *AzureClient, ui packer.Ui) *StepDeleteResourceGroup {
	var step = &StepDeleteResourceGroup{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.delete = step.deleteResourceGroup
	return step
}

func (s *StepDeleteResourceGroup) deleteResourceGroup(resourceGroupName string, cancelCh <-chan struct{}) error {
	_, err := s.client.GroupsClient.Delete(resourceGroupName, cancelCh)

	return err
}

func (s *StepDeleteResourceGroup) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deleting resource group ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))

	result := common.StartInterruptibleTask(
		func() bool { return common.IsStateCancelled(state) },
		func(cancelCh <-chan struct{}) error { return s.delete(resourceGroupName, cancelCh) })

	stepAction := processInterruptibleResult(result, s.error, state)
	state.Put(constants.ArmIsResourceGroupCreated, false)

	return stepAction
}

func (*StepDeleteResourceGroup) Cleanup(multistep.StateBag) {
}
