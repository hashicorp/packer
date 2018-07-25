package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateLoginKeyShouldFailIfOperationCreateLoginKeyFails(t *testing.T) {
	var testSubject = &StepCreateLoginKey{
		CreateLoginKey: func() (*LoginKey, error) { return nil, fmt.Errorf("!! Unit Test FAIL !!") },
		Say:            func(message string) {},
		Error:          func(e error) {},
	}

	stateBag := createTestStateBagStepCreateLoginKey()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepCreateLoginKeyShouldPassIfOperationCreateLoginKeyPasses(t *testing.T) {
	var testSubject = &StepCreateLoginKey{
		CreateLoginKey: func() (*LoginKey, error) { return &LoginKey{"a", "b"}, nil },
		Say:            func(message string) {},
		Error:          func(e error) {},
	}

	stateBag := createTestStateBagStepCreateLoginKey()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateLoginKey() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	return stateBag
}
