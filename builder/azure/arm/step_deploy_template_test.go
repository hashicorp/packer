package arm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepDeployTemplateShouldFailIfDeployFails(t *testing.T) {
	var testSubject = &StepDeployTemplate{
		deploy: func(string, string, <-chan struct{}) error {
			return fmt.Errorf("!! Unit Test FAIL !!")
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepDeployTemplate()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepDeployTemplateShouldPassIfDeployPasses(t *testing.T) {
	var testSubject = &StepDeployTemplate{
		deploy: func(string, string, <-chan struct{}) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepDeployTemplate()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeployTemplateShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	var actualResourceGroupName string
	var actualDeploymentName string

	var testSubject = &StepDeployTemplate{
		deploy: func(resourceGroupName string, deploymentName string, cancelCh <-chan struct{}) error {
			actualResourceGroupName = resourceGroupName
			actualDeploymentName = deploymentName

			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
		name:  "--deployment-name--",
	}

	stateBag := createTestStateBagStepValidateTemplate()
	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)

	if actualDeploymentName != "--deployment-name--" {
		t.Fatal("Expected StepValidateTemplate to source 'constants.ArmDeploymentName' from the state bag, but it did not.")
	}

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatal("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}
}

func TestStepDeployTemplateDeleteImageShouldFailWhenImageUrlCannotBeParsed(t *testing.T) {
	var testSubject = &StepDeployTemplate{
		say:   func(message string) {},
		error: func(e error) {},
		name:  "--deployment-name--",
	}
	// Invalid URL per https://golang.org/src/net/url/url_test.go
	err := testSubject.deleteImage("image", "http://[fe80::1%en0]/", "Unit Test: ResourceGroupName")
	if err == nil {
		t.Fatal("Expected a failure because of the failed image name")
	}
}

func TestStepDeployTemplateDeleteImageShouldFailWithInvalidImage(t *testing.T) {
	var testSubject = &StepDeployTemplate{
		say:   func(message string) {},
		error: func(e error) {},
		name:  "--deployment-name--",
	}
	err := testSubject.deleteImage("image", "storage.blob.core.windows.net/abc", "Unit Test: ResourceGroupName")
	if err == nil {
		t.Fatal("Expected a failure because of the failed image name")
	}
}

func createTestStateBagStepDeployTemplate() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmDeploymentName, "Unit Test: DeploymentName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")

	return stateBag
}
