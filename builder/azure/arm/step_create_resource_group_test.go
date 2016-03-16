// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"testing"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common/constants"
)

func TestStepCreateResourceGroupShouldFailIfCreateFails(t *testing.T) {
	var testSubject = &StepCreateResourceGroup{
		create: func(string, string) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepCreateResourceGroup()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepCreateResourceGroupShouldPassIfCreatePasses(t *testing.T) {
	var testSubject = &StepCreateResourceGroup{
		create: func(string, string) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepCreateResourceGroup()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepCreateResourceGroupShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	var actualResourceGroupName string
	var actualLocation string

	var testSubject = &StepCreateResourceGroup{
		create: func(resourceGroupName string, location string) error {
			actualResourceGroupName = resourceGroupName
			actualLocation = location
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepCreateResourceGroup()
	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	var expectedLocation = stateBag.Get(constants.ArmLocation).(string)
	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatalf("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	if actualLocation != expectedLocation {
		t.Fatalf("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}
}

func createTestStateBagStepCreateResourceGroup() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmLocation, "Unit Test: Location")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")

	return stateBag
}
