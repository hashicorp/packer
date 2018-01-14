package ncloud

import (
	"fmt"
	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepDeletePublicIPInstanceShouldFailIfOperationDeletePublicIPInstanceFails(t *testing.T) {
	var testSubject = &StepDeletePublicIPInstance{
		DeletePublicIPInstance: func(publicIPInstanceNo string) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		Say:   func(message string) {},
		Error: func(e error) {},
	}

	stateBag := createTestStateBagStepDeletePublicIPInstance()

	var result = testSubject.Run(stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepDeletePublicIPInstanceShouldPassIfOperationDeletePublicIPInstancePasses(t *testing.T) {
	var testSubject = &StepDeletePublicIPInstance{
		DeletePublicIPInstance: func(publicIPInstanceNo string) error { return nil },
		Say:   func(message string) {},
		Error: func(e error) {},
	}

	stateBag := createTestStateBagStepDeletePublicIPInstance()

	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepDeletePublicIPInstance() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("PublicIPInstance", &ncloud.PublicIPInstance{PublicIPInstanceNo: "22"})

	return stateBag
}
