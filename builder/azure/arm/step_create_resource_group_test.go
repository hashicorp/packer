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
		create: func(string, string, *map[string]*string) error { return fmt.Errorf("!! Unit Test FAIL !!") },
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
		create: func(string, string, *map[string]*string) error { return nil },
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
	var actualTags *map[string]*string

	var testSubject = &StepCreateResourceGroup{
		create: func(resourceGroupName string, location string, tags *map[string]*string) error {
			actualResourceGroupName = resourceGroupName
			actualLocation = location
			actualTags = tags
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

	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)
	var expectedLocation = stateBag.Get(constants.ArmLocation).(string)
	var expectedTags = stateBag.Get(constants.ArmTags).(*map[string]*string)

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatal("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	if actualLocation != expectedLocation {
		t.Fatal("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	if len(*expectedTags) != len(*actualTags) && *(*expectedTags)["tag01"] != *(*actualTags)["tag01"] {
		t.Fatal("Expected the step to source 'constants.ArmTags' from the state bag, but it did not.")
	}

	_, ok := stateBag.GetOk(constants.ArmIsResourceGroupCreated)
	if !ok {
		t.Fatal("Expected the step to add item to stateBag['constants.ArmIsResourceGroupCreated'], but it did not.")
	}
}

func createTestStateBagStepCreateResourceGroup() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmLocation, "Unit Test: Location")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")

	value := "Unit Test: Tags"
	tags := map[string]*string{
		"tag01": &value,
	}

	stateBag.Put(constants.ArmTags, &tags)

	return stateBag
}
