package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepCreateBlockStorageInstanceShouldFailIfOperationCreateBlockStorageInstanceFails(t *testing.T) {

	var testSubject = &StepCreateBlockStorage{
		CreateBlockStorage: func(serverInstanceNo string) (*string, error) { return nil, fmt.Errorf("!! Unit Test FAIL !!") },
		Say:                func(message string) {},
		Error:              func(e error) {},
		Config:             new(Config),
	}

	testSubject.Config.BlockStorageSize = 10

	stateBag := createTestStateBagStepCreateBlockStorageInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepCreateBlockStorageInstanceShouldPassIfOperationCreateBlockStorageInstancePasses(t *testing.T) {
	var instanceNo = "a"
	var testSubject = &StepCreateBlockStorage{
		CreateBlockStorage: func(serverInstanceNo string) (*string, error) { return &instanceNo, nil },
		Say:                func(message string) {},
		Error:              func(e error) {},
		Config:             new(Config),
	}

	testSubject.Config.BlockStorageSize = 10

	stateBag := createTestStateBagStepCreateBlockStorageInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateBlockStorageInstance() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("instance_no", "a")

	return stateBag
}
