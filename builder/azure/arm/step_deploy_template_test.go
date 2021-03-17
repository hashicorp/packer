package arm

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/azure/common/constants"
)

func TestStepDeployTemplateShouldFailIfDeployFails(t *testing.T) {
	var testSubject = &StepDeployTemplate{
		deploy: func(context.Context, string, string) error {
			return fmt.Errorf("!! Unit Test FAIL !!")
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepDeployTemplate()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepDeployTemplateShouldPassIfDeployPasses(t *testing.T) {
	var testSubject = &StepDeployTemplate{
		deploy: func(context.Context, string, string) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := createTestStateBagStepDeployTemplate()

	var result = testSubject.Run(context.Background(), stateBag)
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
		deploy: func(ctx context.Context, resourceGroupName string, deploymentName string) error {
			actualResourceGroupName = resourceGroupName
			actualDeploymentName = deploymentName

			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
		name:  "--deployment-name--",
	}

	stateBag := createTestStateBagStepValidateTemplate()
	var result = testSubject.Run(context.Background(), stateBag)

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
	err := testSubject.deleteImage(context.TODO(), "image", "http://[fe80::1%en0]/", "Unit Test: ResourceGroupName")
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
	err := testSubject.deleteImage(context.TODO(), "image", "storage.blob.core.windows.net/abc", "Unit Test: ResourceGroupName")
	if err == nil {
		t.Fatal("Expected a failure because of the failed image name")
	}
}

func TestStepDeployTemplateCleanupShouldDeleteManagedOSImageInExistingResourceGroup(t *testing.T) {
	var deleteDiskCounter = 0
	var testSubject = createTestStepDeployTemplateDeleteOSImage(&deleteDiskCounter)

	stateBag := createTestStateBagStepDeployTemplate()
	stateBag.Put(constants.ArmIsManagedImage, true)
	stateBag.Put(constants.ArmIsExistingResourceGroup, true)
	stateBag.Put(constants.ArmIsResourceGroupCreated, true)
	stateBag.Put("ui", packersdk.TestUi(t))

	testSubject.Cleanup(stateBag)
	if deleteDiskCounter != 1 {
		t.Fatalf("Expected DeployTemplate Cleanup to invoke deleteDisk 1 time, but invoked %d times", deleteDiskCounter)
	}
}

func TestStepDeployTemplateCleanupShouldDeleteManagedOSImageInTemporaryResourceGroup(t *testing.T) {
	var deleteDiskCounter = 0
	var testSubject = createTestStepDeployTemplateDeleteOSImage(&deleteDiskCounter)

	stateBag := createTestStateBagStepDeployTemplate()
	stateBag.Put(constants.ArmIsManagedImage, true)
	stateBag.Put(constants.ArmIsExistingResourceGroup, false)
	stateBag.Put(constants.ArmIsResourceGroupCreated, true)
	stateBag.Put("ui", packersdk.TestUi(t))

	testSubject.Cleanup(stateBag)
	if deleteDiskCounter != 1 {
		t.Fatalf("Expected DeployTemplate Cleanup to invoke deleteDisk 1 times, but invoked %d times", deleteDiskCounter)
	}
}

func TestStepDeployTemplateCleanupShouldDeleteVHDOSImageInExistingResourceGroup(t *testing.T) {
	var deleteDiskCounter = 0
	var testSubject = createTestStepDeployTemplateDeleteOSImage(&deleteDiskCounter)

	stateBag := createTestStateBagStepDeployTemplate()
	stateBag.Put(constants.ArmIsManagedImage, false)
	stateBag.Put(constants.ArmIsExistingResourceGroup, true)
	stateBag.Put(constants.ArmIsResourceGroupCreated, true)
	stateBag.Put("ui", packersdk.TestUi(t))

	testSubject.Cleanup(stateBag)
	if deleteDiskCounter != 1 {
		t.Fatalf("Expected DeployTemplate Cleanup to invoke deleteDisk 1 time, but invoked %d times", deleteDiskCounter)
	}
}

func TestStepDeployTemplateCleanupShouldVHDOSImageInTemporaryResourceGroup(t *testing.T) {
	var deleteDiskCounter = 0
	var testSubject = createTestStepDeployTemplateDeleteOSImage(&deleteDiskCounter)

	stateBag := createTestStateBagStepDeployTemplate()
	stateBag.Put(constants.ArmIsManagedImage, false)
	stateBag.Put(constants.ArmIsExistingResourceGroup, false)
	stateBag.Put(constants.ArmIsResourceGroupCreated, true)
	stateBag.Put("ui", packersdk.TestUi(t))

	testSubject.Cleanup(stateBag)
	if deleteDiskCounter != 1 {
		t.Fatalf("Expected DeployTemplate Cleanup to invoke deleteDisk 1 times, but invoked %d times", deleteDiskCounter)
	}
}

func createTestStateBagStepDeployTemplate() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmDeploymentName, "Unit Test: DeploymentName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")
	stateBag.Put(constants.ArmComputeName, "Unit Test: ComputeName")

	return stateBag
}

func createTestStepDeployTemplateDeleteOSImage(deleteDiskCounter *int) *StepDeployTemplate {
	return &StepDeployTemplate{
		deploy: func(context.Context, string, string) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
		deleteDisk: func(ctx context.Context, imageType string, imageName string, resourceGroupName string) error {
			*deleteDiskCounter++
			return nil
		},
		disk: func(ctx context.Context, resourceGroupName, computeName string) (string, string, error) {
			return "Microsoft.Compute/disks", "", nil
		},
		delete: func(ctx context.Context, deploymentName, resourceGroupName string) error {
			return nil
		},
		deleteDeployment: func(ctx context.Context, state multistep.StateBag) error {
			return nil
		},
	}
}
