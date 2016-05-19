// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/packer"
)

type StepCaptureImage struct {
	client  *AzureClient
	capture func(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters, cancelCh <-chan struct{}) error
	get     func(client *AzureClient) *CaptureTemplate
	say     func(message string)
	error   func(e error)
}

func NewStepCaptureImage(client *AzureClient, ui packer.Ui) *StepCaptureImage {
	var step = &StepCaptureImage{
		client: client,
		get:    func(client *AzureClient) *CaptureTemplate { return client.Template },
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.capture = step.captureImage
	return step
}

func (s *StepCaptureImage) captureImage(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters, cancelCh <-chan struct{}) error {
	_, err := s.client.Generalize(resourceGroupName, computeName)
	if err != nil {
		return err
	}

	_, err = s.client.Capture(resourceGroupName, computeName, *parameters, cancelCh)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepCaptureImage) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Capturing image ...")

	var computeName = state.Get(constants.ArmComputeName).(string)
	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var parameters = state.Get(constants.ArmVirtualMachineCaptureParameters).(*compute.VirtualMachineCaptureParameters)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	result := common.StartInterruptibleTask(
		func() bool { return common.IsStateCancelled(state) },
		func(cancelCh <-chan struct{}) error {
			return s.capture(resourceGroupName, computeName, parameters, cancelCh)
		})

	// HACK(chrboum): I do not like this.  The capture method should be returning this value
	// instead having to pass in another lambda.  I'm in this pickle because I am using
	// common.StartInterruptibleTask which is not parametric, and only returns a type of error.
	// I could change it to interface{}, but I do not like that solution either.
	//
	// Having to resort to capturing the template via an inspector is hack, and once I can
	// resolve that I can cleanup this code too.  See the comments in azure_client.go for more
	// details.
	template := s.get(s.client)
	state.Put(constants.ArmCaptureTemplate, template)

	return processInterruptibleResult(result, s.error, state)
}

func (*StepCaptureImage) Cleanup(multistep.StateBag) {
}
