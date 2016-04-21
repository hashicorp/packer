// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepGetIPAddress struct {
	client *AzureClient
	get    func(resourceGroupName string, ipAddressName string) (string, error)
	say    func(message string)
	error  func(e error)
}

func NewStepGetIPAddress(client *AzureClient, ui packer.Ui) *StepGetIPAddress {
	var step = &StepGetIPAddress{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.get = step.getIPAddress
	return step
}

func (s *StepGetIPAddress) getIPAddress(resourceGroupName string, ipAddressName string) (string, error) {
	res, err := s.client.PublicIPAddressesClient.Get(resourceGroupName, ipAddressName, "")
	if err != nil {
		return "", nil
	}

	return *res.Properties.IPAddress, nil
}

func (s *StepGetIPAddress) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Getting the public IP address ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var ipAddressName = state.Get(constants.ArmPublicIPAddressName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName   : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> PublicIPAddressName : '%s'", ipAddressName))

	address, err := s.get(resourceGroupName, ipAddressName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	s.say(fmt.Sprintf(" -> Public IP           : '%s'", address))
	state.Put(constants.SSHHost, address)

	return multistep.ActionContinue
}

func (*StepGetIPAddress) Cleanup(multistep.StateBag) {
}
