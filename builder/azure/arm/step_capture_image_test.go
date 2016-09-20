// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common/constants"
)

func TestStepCaptureImageShouldFailIfCaptureFails(t *testing.T) {
	var testSubject = &StepCaptureImage{
		capture: func(string, string, *compute.VirtualMachineCaptureParameters, <-chan struct{}) error {
			return fmt.Errorf("!! Unit Test FAIL !!")
		},
		get: func(client *AzureClient) *CaptureTemplate {
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepCaptureImage()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepCaptureImageShouldPassIfCapturePasses(t *testing.T) {
	var testSubject = &StepCaptureImage{
		capture: func(string, string, *compute.VirtualMachineCaptureParameters, <-chan struct{}) error { return nil },
		get: func(client *AzureClient) *CaptureTemplate {
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepCaptureImage()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepCaptureImageShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	cancelCh := make(chan<- struct{})
	defer close(cancelCh)

	var actualResourceGroupName string
	var actualComputeName string
	var actualVirtualMachineCaptureParameters *compute.VirtualMachineCaptureParameters
	actualCaptureTemplate := &CaptureTemplate{
		Schema: "!! Unit Test !!",
	}

	var testSubject = &StepCaptureImage{
		capture: func(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters, cancelCh <-chan struct{}) error {
			actualResourceGroupName = resourceGroupName
			actualComputeName = computeName
			actualVirtualMachineCaptureParameters = parameters

			return nil
		},
		get: func(client *AzureClient) *CaptureTemplate {
			return actualCaptureTemplate
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepCaptureImage()
	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	var expectedComputeName = stateBag.Get(constants.ArmComputeName).(string)
	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)
	var expectedVirtualMachineCaptureParameters = stateBag.Get(constants.ArmVirtualMachineCaptureParameters).(*compute.VirtualMachineCaptureParameters)
	var expectedCaptureTemplate = stateBag.Get(constants.ArmCaptureTemplate).(*CaptureTemplate)

	if actualComputeName != expectedComputeName {
		t.Fatal("Expected StepCaptureImage to source 'constants.ArmComputeName' from the state bag, but it did not.")
	}

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatal("Expected StepCaptureImage to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	if actualVirtualMachineCaptureParameters != expectedVirtualMachineCaptureParameters {
		t.Fatal("Expected StepCaptureImage to source 'constants.ArmVirtualMachineCaptureParameters' from the state bag, but it did not.")
	}

	if actualCaptureTemplate != expectedCaptureTemplate {
		t.Fatal("Expected StepCaptureImage to source 'constants.ArmCaptureTemplate' from the state bag, but it did not.")
	}
}

func createTestStateBagStepCaptureImage() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmComputeName, "Unit Test: ComputeName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")
	stateBag.Put(constants.ArmVirtualMachineCaptureParameters, &compute.VirtualMachineCaptureParameters{})

	return stateBag
}
