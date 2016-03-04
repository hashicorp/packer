// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCreateResourceGroup struct {
	client *AzureClient
	create func(resourceGroupName string, location string) error
	say    func(message string)
	error  func(e error)
}

func NewStepCreateResourceGroup(client *AzureClient, ui packer.Ui) *StepCreateResourceGroup {
	var step = &StepCreateResourceGroup{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.create = step.createResourceGroup
	return step
}

func (s *StepCreateResourceGroup) createResourceGroup(resourceGroupName string, location string) error {
	_, err := s.client.GroupsClient.CreateOrUpdate(resourceGroupName, resources.ResourceGroup{
		Location: &location,
	})

	return err
}

func (s *StepCreateResourceGroup) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Creating resource group ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var location = state.Get(constants.ArmLocation).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> Location          : '%s'", location))

	err := s.create(resourceGroupName, location)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepCreateResourceGroup) Cleanup(multistep.StateBag) {
}
