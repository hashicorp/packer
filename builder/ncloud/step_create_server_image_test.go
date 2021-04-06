package ncloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepCreateServerImageShouldFailIfOperationCreateServerImageFails(t *testing.T) {
	var testSubject = &StepCreateServerImage{
		CreateServerImage: func(serverInstanceNo string) (*server.MemberServerImage, error) {
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

	if _, ok := stateBag.GetOk("error"); ok == false {
		t.Fatal("Expected the step to set stateBag['Error'], but it was not.")
	}
}
func TestStepCreateServerImageShouldPassIfOperationCreateServerImagePasses(t *testing.T) {
	var testSubject = &StepCreateServerImage{
		CreateServerImage: func(serverInstanceNo string) (*server.MemberServerImage, error) { return nil, nil },
		Say:               func(message string) {},
		Error:             func(e error) {},
	}

	stateBag := createTestStateBagStepCreateServerImage()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateServerImage() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("instance_no", "a")

	return stateBag
}
