package arm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
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
	get      func(ctx context.Context, resourceGroupName string, ipAddressName string, interfaceName string) (string, error)
	say      func(message string)
	error    func(e error)
}

func NewStepGetIPAddress(client *AzureClient, ui packersdk.Ui, endpoint EndpointType) *StepGetIPAddress {
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

func (s *StepGetIPAddress) getPrivateIP(ctx context.Context, resourceGroupName string, ipAddressName string, interfaceName string) (string, error) {
	resp, err := s.client.InterfacesClient.Get(ctx, resourceGroupName, interfaceName, "")
	if err != nil {
		s.say(s.client.LastError.Error())
		return "", err
	}

	return *(*resp.IPConfigurations)[0].PrivateIPAddress, nil
}

func (s *StepGetIPAddress) getPublicIP(ctx context.Context, resourceGroupName string, ipAddressName string, interfaceName string) (string, error) {
	resp, err := s.client.PublicIPAddressesClient.Get(ctx, resourceGroupName, ipAddressName, "")
	if err != nil {
		return "", err
	}

	return *resp.IPAddress, nil
}

func (s *StepGetIPAddress) getPublicIPInPrivateNetwork(ctx context.Context, resourceGroupName string, ipAddressName string, interfaceName string) (string, error) {
	s.getPrivateIP(ctx, resourceGroupName, ipAddressName, interfaceName)
	return s.getPublicIP(ctx, resourceGroupName, ipAddressName, interfaceName)
}

func (s *StepGetIPAddress) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Getting the VM's IP address ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var ipAddressName = state.Get(constants.ArmPublicIPAddressName).(string)
	var nicName = state.Get(constants.ArmNicName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName   : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> PublicIPAddressName : '%s'", ipAddressName))
	s.say(fmt.Sprintf(" -> NicName             : '%s'", nicName))
	s.say(fmt.Sprintf(" -> Network Connection  : '%s'", EndpointCommunicationText[s.endpoint]))

	address, err := s.get(ctx, resourceGroupName, ipAddressName, nicName)
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
