package ncloud

import (
	"context"
	"fmt"
	"testing"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreatePublicIPInstanceShouldFailIfOperationCreatePublicIPInstanceFails(t *testing.T) {
	var testSubject = &StepCreatePublicIPInstance{
		CreatePublicIPInstance: func(serverInstanceNo string) (*ncloud.PublicIPInstance, error) {
			return nil, fmt.Errorf("!! Unit Test FAIL !!")
		},
		Say:   func(message string) {},
		Error: func(e error) {},
	}

	stateBag := createTestStateBagStepCreateServerImage()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}

func TestStepCreatePublicIPInstanceShouldPassIfOperationCreatePublicIPInstancePasses(t *testing.T) {
	c := new(Config)
	c.Comm.Prepare(nil)
	c.Comm.Type = "ssh"

	var testSubject = &StepCreatePublicIPInstance{
		CreatePublicIPInstance: func(serverInstanceNo string) (*ncloud.PublicIPInstance, error) {
			return &ncloud.PublicIPInstance{PublicIPInstanceNo: "a", PublicIP: "b"}, nil
		},
		Say:    func(message string) {},
		Error:  func(e error) {},
		Config: c,
	}

	stateBag := createTestStateBagStepCreatePublicIPInstance()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreatePublicIPInstance() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("InstanceNo", "a")

	return stateBag
}
