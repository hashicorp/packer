package ncloud

import (
	"context"
	"fmt"
	"testing"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateServerImageShouldFailIfOperationCreateServerImageFails(t *testing.T) {
	var testSubject = &StepCreateServerImage{
		CreateServerImage: func(serverInstanceNo string) (*ncloud.ServerImage, error) {
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
func TestStepCreateServerImageShouldPassIfOperationCreateServerImagePasses(t *testing.T) {
	var testSubject = &StepCreateServerImage{
		CreateServerImage: func(serverInstanceNo string) (*ncloud.ServerImage, error) { return nil, nil },
		Say:               func(message string) {},
		Error:             func(e error) {},
	}

	stateBag := createTestStateBagStepCreateServerImage()

	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk("Error"); ok == true {
		t.Fatalf("Expected the step to not set stateBag['Error'], but it was.")
	}
}

func createTestStateBagStepCreateServerImage() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put("InstanceNo", "a")

	return stateBag
}
