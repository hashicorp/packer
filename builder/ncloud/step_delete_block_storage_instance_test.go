package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepDeleteBlockStorageInstanceShouldFailIfOperationDeleteBlockStorageInstanceFails(t *testing.T) {
	var testSubject = &StepDeleteBlockStorageInstance{
		DeleteBlockStorageInstance: func(blockStorageInstanceNo string) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		Say:    func(message string) {},
		Error:  func(e error) {},
		Config: &Config{BlockStorageSize: 10},
	}

	stateBag := createTestStateBagStepDeleteBlockStorageInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepDeleteBlockStorageInstanceShouldPassIfOperationDeleteBlockStorageInstancePasses(t *testing.T) {
	var testSubject = &StepDeleteBlockStorageInstance{
		DeleteBlockStorageInstance: func(blockStorageInstanceNo string) error { return nil },
		Say:    func(message string) {},
		Error:  func(e error) {},
		Config: &Config{BlockStorageSize: 10},
	}

	stateBag := createTestStateBagStepDeleteBlockStorageInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepDeleteBlockStorageInstance() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("InstanceNo", "1")

	return stateBag
}
