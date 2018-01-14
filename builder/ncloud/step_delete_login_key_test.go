package ncloud

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepDeleteLoginKeyShouldFailIfOperationDeleteLoginKeyFails(t *testing.T) {
	var testSubject = &StepDeleteLoginKey{
		DeleteLoginKey: func(keyName string) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		Say:            func(message string) {},
		Error:          func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteLoginKey()

	var result = testSubject.Run(stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepDeleteLoginKeyShouldPassIfOperationDeleteLoginKeyPasses(t *testing.T) {
	var testSubject = &StepDeleteLoginKey{
		DeleteLoginKey: func(keyName string) error { return nil },
		Say:            func(message string) {},
		Error:          func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteLoginKey()

	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func DeleteTestStateBagStepDeleteLoginKey() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("LoginKey", &LoginKey{"a", "b"})

	return stateBag
}
