package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepCreateServerInstanceShouldFailIfOperationCreateFails(t *testing.T) {
	var testSubject = &StepCreateServerInstance{
		CreateServerInstance: func(loginKeyName string, feeSystemTypeCode string, state multistep.StateBag) (string, error) {
			return "", fmt.Errorf("!! Unit Test FAIL !!")
		},
		Say:   func(message string) {},
		Error: func(e error) {},
	}

	stateBag := createTestStateBagStepCreateServerInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepCreateServerInstanceShouldPassIfOperationCreatePasses(t *testing.T) {
	var testSubject = &StepCreateServerInstance{
		CreateServerInstance: func(loginKeyName string, feeSystemTypeCode string, state multistep.StateBag) (string, error) {
			return "", nil
		},
		Say:   func(message string) {},
		Error: func(e error) {},
	}

	stateBag := createTestStateBagStepCreateServerInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateServerInstance() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("login_key", &LoginKey{"a", "b"})
	stateBag.Put("zone_no", "1")

	return stateBag
}
