package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepCreateInitScriptShouldFailIfOperationCreateInitScriptFails(t *testing.T) {
	var testSubject = &StepCreateInitScript{
		CreateInitScript: func() (string, error) { return "", fmt.Errorf("!! Unit Test FAIL !!") },
		Say:              func(message string) {},
		Error:            func(e error) {},
		Config: &Config{
			Region:     "Korea",
			SupportVPC: true,
			UserData:   "test",
		},
	}

	stateBag := createTestStateBagStepCreateInitScript()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepCreateInitScriptShouldPassIfOperationCreateInitScriptPasses(t *testing.T) {
	var testSubject = &StepCreateInitScript{
		CreateInitScript: func() (string, error) { return "123", nil },
		Say:              func(message string) {},
		Error:            func(e error) {},
		Config: &Config{
			Region:     "Korea",
			SupportVPC: true,
		},
	}

	stateBag := createTestStateBagStepCreateInitScript()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateInitScript() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	return stateBag
}
