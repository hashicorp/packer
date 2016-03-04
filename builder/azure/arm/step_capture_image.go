// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCaptureImage struct {
	client  *AzureClient
	capture func(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters) error
	say     func(message string)
	error   func(e error)
}

func NewStepCaptureImage(client *AzureClient, ui packer.Ui) *StepCaptureImage {
	var step = &StepCaptureImage{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.capture = step.captureImage
	return step
}

func (s *StepCaptureImage) captureImage(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters) error {
	generalizeResponse, err := s.client.Generalize(resourceGroupName, computeName)
	if err != nil {
		return err
	}

	s.client.VirtualMachinesClient.PollAsNeeded(generalizeResponse.Response)

	captureResponse, err := s.client.Capture(resourceGroupName, computeName, *parameters)
	if err != nil {
		return err
	}

	s.client.VirtualMachinesClient.PollAsNeeded(captureResponse.Response.Response)
	return nil
}

func (s *StepCaptureImage) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Capturing image ...")

	var computeName = state.Get(constants.ArmComputeName).(string)
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var parameters = state.Get(constants.ArmVirtualMachineCaptureParameters).(*compute.VirtualMachineCaptureParameters)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	err := s.capture(resourceGroupName, computeName, parameters)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepCaptureImage) Cleanup(multistep.StateBag) {
}
