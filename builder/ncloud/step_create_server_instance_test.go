package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateServerInstanceShouldFailIfOperationCreateFails(t *testing.T) {
	var testSubject = &StepCreateServerInstance{
		CreateServerInstance: func(loginKeyName string, zoneNo string, feeSystemTypeCode string) (string, error) {
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

	if _, ok := stateBag.GetOk("Error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepCreateServerInstanceShouldPassIfOperationCreatePasses(t *testing.T) {
	var testSubject = &StepCreateServerInstance{
		CreateServerInstance: func(loginKeyName string, zoneNo string, feeSystemTypeCode string) (string, error) { return "", nil },
		Say:                  func(message string) {},
		Error:                func(e error) {},
	}

	stateBag := createTestStateBagStepCreateServerInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateServerInstance() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("LoginKey", &LoginKey{"a", "b"})
	stateBag.Put("ZoneNo", "1")

	return stateBag
}
