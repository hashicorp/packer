// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"testing"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common/constants"
)

func TestStepDeleteResourceGroupShouldFailIfDeleteFails(t *testing.T) {
	var testSubject = &StepDeleteResourceGroup{
		delete: func(string, <-chan struct{}) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteResourceGroup()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepDeleteResourceGroupShouldPassIfDeletePasses(t *testing.T) {
	var testSubject = &StepDeleteResourceGroup{
		delete: func(string, <-chan struct{}) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteResourceGroup()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeleteResourceGroupShouldDeleteStateBagArmResourceGroupCreated(t *testing.T) {
	var testSubject = &StepDeleteResourceGroup{
		delete: func(resourceGroupName string, cancelCh <-chan struct{}) error {
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteResourceGroup()
	testSubject.Run(stateBag)

	value, ok := stateBag.GetOk(constants.ArmIsResourceGroupCreated)
	if !ok {
		t.Fatal("Expected the resource bag value arm.IsResourceGroupCreated to exist")
	}

	if value.(bool) {
		t.Fatalf("Expected arm.IsResourceGroupCreated to be false, but got %q", value)
	}
}

func DeleteTestStateBagStepDeleteResourceGroup() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")
	stateBag.Put(constants.ArmIsResourceGroupCreated, "Unit Test: IsResourceGroupCreated")

	return stateBag
}
