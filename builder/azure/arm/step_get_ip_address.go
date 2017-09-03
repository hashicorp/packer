// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type EndpointType int

const (
	PublicEndpoint EndpointType = iota
	PrivateEndpoint
	PublicEndpointInPrivateNetwork
)

var (
	EndpointCommunicationText = map[EndpointType]string{
		PublicEndpoint:                 "PublicEndpoint",
		PrivateEndpoint:                "PrivateEndpoint",
		PublicEndpointInPrivateNetwork: "PublicEndpointInPrivateNetwork",
	}
)

type StepGetIPAddress struct {
	client   *AzureClient
	endpoint EndpointType
	get      func(resourceGroupName string, ipAddressName string, interfaceName string) (string, error)
	say      func(message string)
	error    func(e error)
}

func NewStepGetIPAddress(client *AzureClient, ui packer.Ui, endpoint EndpointType) *StepGetIPAddress {
	var step = &StepGetIPAddress{
		client:   client,
		endpoint: endpoint,
		say:      func(message string) { ui.Say(message) },
		error:    func(e error) { ui.Error(e.Error()) },
	}

	switch endpoint {
	case PrivateEndpoint:
		step.get = step.getPrivateIP
	case PublicEndpoint:
		step.get = step.getPublicIP
	case PublicEndpointInPrivateNetwork:
		step.get = step.getPublicIPInPrivateNetwork
	}

	return step
}

func (s *StepGetIPAddress) getPrivateIP(resourceGroupName string, ipAddressName string, interfaceName string) (string, error) {
	resp, err := s.client.InterfacesClient.Get(resourceGroupName, interfaceName, "")
	if err != nil {
		s.say(s.client.LastError.Error())
		return "", err
	}

	return *(*resp.IPConfigurations)[0].PrivateIPAddress, nil
}

func (s *StepGetIPAddress) getPublicIP(resourceGroupName string, ipAddressName string, interfaceName string) (string, error) {
	resp, err := s.client.PublicIPAddressesClient.Get(resourceGroupName, ipAddressName, "")
	if err != nil {
		return "", err
	}

	return *resp.IPAddress, nil
}

func (s *StepGetIPAddress) getPublicIPInPrivateNetwork(resourceGroupName string, ipAddressName string, interfaceName string) (string, error) {
	s.getPrivateIP(resourceGroupName, ipAddressName, interfaceName)
	return s.getPublicIP(resourceGroupName, ipAddressName, interfaceName)
}

func (s *StepGetIPAddress) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Getting the VM's IP address ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var ipAddressName = state.Get(constants.ArmPublicIPAddressName).(string)
	var nicName = state.Get(constants.ArmNicName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName   : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> PublicIPAddressName : '%s'", ipAddressName))
	s.say(fmt.Sprintf(" -> NicName             : '%s'", nicName))
	s.say(fmt.Sprintf(" -> Network Connection  : '%s'", EndpointCommunicationText[s.endpoint]))

	address, err := s.get(resourceGroupName, ipAddressName, nicName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	state.Put(constants.SSHHost, address)
	s.say(fmt.Sprintf(" -> IP Address          : '%s'", address))

	return multistep.ActionContinue
}

func (*StepGetIPAddress) Cleanup(multistep.StateBag) {
}
