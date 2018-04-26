package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateBlockStorageInstanceShouldFailIfOperationCreateBlockStorageInstanceFails(t *testing.T) {

	var testSubject = &StepCreateBlockStorageInstance{
		CreateBlockStorageInstance: func(serverInstanceNo string) (string, error) { return "", fmt.Errorf("!! Unit Test FAIL !!") },
		Say:    func(message string) {},
		Error:  func(e error) {},
		Config: new(Config),
	}

	testSubject.Config.BlockStorageSize = 10

	stateBag := createTestStateBagStepCreateBlockStorageInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepCreateBlockStorageInstanceShouldPassIfOperationCreateBlockStorageInstancePasses(t *testing.T) {
	var testSubject = &StepCreateBlockStorageInstance{
		CreateBlockStorageInstance: func(serverInstanceNo string) (string, error) { return "a", nil },
		Say:    func(message string) {},
		Error:  func(e error) {},
		Config: new(Config),
	}

	testSubject.Config.BlockStorageSize = 10

	stateBag := createTestStateBagStepCreateBlockStorageInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateBlockStorageInstance() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("InstanceNo", "a")

	return stateBag
}
