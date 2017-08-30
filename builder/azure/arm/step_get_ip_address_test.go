// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
)

func TestStepGetIPAddressShouldFailIfGetFails(t *testing.T) {
	endpoints := []EndpointType{PublicEndpoint, PublicEndpointInPrivateNetwork}

	for _, endpoint := range endpoints {
		var testSubject = &StepGetIPAddress{
			get:      func(string, string, string) (string, error) { return "", fmt.Errorf("!! Unit Test FAIL !!") },
			endpoint: endpoint,
			say:      func(message string) {},
			error:    func(e error) {},
		}

		stateBag := createTestStateBagStepGetIPAddress()

		var result = testSubject.Run(stateBag)
		if result != multistep.ActionHalt {
			t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
		}

		if _, ok := stateBag.GetOk(constants.Error); ok == false {
			t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
		}
	}
}

func TestStepGetIPAddressShouldPassIfGetPasses(t *testing.T) {
	endpoints := []EndpointType{PublicEndpoint, PublicEndpointInPrivateNetwork}

	for _, endpoint := range endpoints {
		var testSubject = &StepGetIPAddress{
			get:      func(string, string, string) (string, error) { return "", nil },
			endpoint: endpoint,
			say:      func(message string) {},
			error:    func(e error) {},
		}

		stateBag := createTestStateBagStepGetIPAddress()

		var result = testSubject.Run(stateBag)
		if result != multistep.ActionContinue {
			t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
		}

		if _, ok := stateBag.GetOk(constants.Error); ok == true {
			t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
		}
	}
}

func TestStepGetIPAddressShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	var actualResourceGroupName string
	var actualIPAddressName string
	var actualNicName string
	endpoints := []EndpointType{PublicEndpoint, PublicEndpointInPrivateNetwork}

	for _, endpoint := range endpoints {
		var testSubject = &StepGetIPAddress{
			get: func(resourceGroupName string, ipAddressName string, nicName string) (string, error) {
				actualResourceGroupName = resourceGroupName
				actualIPAddressName = ipAddressName
				actualNicName = nicName

				return "127.0.0.1", nil
			},
			endpoint: endpoint,
			say:      func(message string) {},
			error:    func(e error) {},
		}

		stateBag := createTestStateBagStepGetIPAddress()
		var result = testSubject.Run(stateBag)

		if result != multistep.ActionContinue {
			t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
		}

		var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)
		var expectedIPAddressName = stateBag.Get(constants.ArmPublicIPAddressName).(string)
		var expectedNicName = stateBag.Get(constants.ArmNicName).(string)

		if actualIPAddressName != expectedIPAddressName {
			t.Fatal("Expected StepGetIPAddress to source 'constants.ArmIPAddressName' from the state bag, but it did not.")
		}

		if actualResourceGroupName != expectedResourceGroupName {
			t.Fatal("Expected StepGetIPAddress to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
		}

		if actualNicName != expectedNicName {
			t.Fatalf("Expected StepGetIPAddress to source 'constants.ArmNetworkInterfaceName' from the state bag, but it did not.")
		}

		expectedIPAddress, ok := stateBag.GetOk(constants.SSHHost)
		if !ok {
			t.Fatalf("Expected the state bag to have a value for '%s', but it did not.", constants.SSHHost)
		}

		if expectedIPAddress != "127.0.0.1" {
			t.Fatalf("Expected the value of stateBag[%s] to be '127.0.0.1', but got '%s'.", constants.SSHHost, expectedIPAddress)
		}
	}
}

func createTestStateBagStepGetIPAddress() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmPublicIPAddressName, "Unit Test: PublicIPAddressName")
	stateBag.Put(constants.ArmNicName, "Unit Test: NicName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")

	return stateBag
}
