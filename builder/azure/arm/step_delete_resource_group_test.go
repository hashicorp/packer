package arm

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepDeleteResourceGroupShouldFailIfDeleteFails(t *testing.T) {
	var testSubject = &StepDeleteResourceGroup{
		delete: func(multistep.StateBag, string, <-chan struct{}) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteResourceGroup()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepDeleteResourceGroupShouldPassIfDeletePasses(t *testing.T) {
	var testSubject = &StepDeleteResourceGroup{
		delete: func(multistep.StateBag, string, <-chan struct{}) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteResourceGroup()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeleteResourceGroupShouldDeleteStateBagArmResourceGroupCreated(t *testing.T) {
	var testSubject = &StepDeleteResourceGroup{
		delete: func(s multistep.StateBag, resourceGroupName string, cancelCh <-chan struct{}) error {
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteResourceGroup()
	testSubject.Run(context.Background(), stateBag)

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
