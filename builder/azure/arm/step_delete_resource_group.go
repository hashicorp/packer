// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepDeleteResourceGroup struct {
	client *AzureClient
	delete func(resourceGroupName string) error
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

func (s *StepDeleteResourceGroup) deleteResourceGroup(resourceGroupName string) error {
	res, err := s.client.GroupsClient.Delete(resourceGroupName)
	if err != nil {
		return err
	}

	s.client.GroupsClient.PollAsNeeded(res.Response)
	return nil
}

func (s *StepDeleteResourceGroup) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deleting resource group ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))

	err := s.delete(resourceGroupName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepDeleteResourceGroup) Cleanup(multistep.StateBag) {
}
