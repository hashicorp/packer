package arm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepDeleteAdditionalDiskShouldFailIfGetFails(t *testing.T) {
	var testSubject = &StepDeleteAdditionalDisk{
		delete:        func(string, string) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		deleteManaged: func(context.Context, string, string) error { return nil },
		say:           func(message string) {},
		error:         func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteAdditionalDisk([]string{"http://storage.blob.core.windows.net/images/pkrvm_os.vhd"})

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepDeleteAdditionalDiskShouldPassIfGetPasses(t *testing.T) {
	var testSubject = &StepDeleteAdditionalDisk{
		delete: func(string, string) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteAdditionalDisk([]string{"http://storage.blob.core.windows.net/images/pkrvm_os.vhd"})

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeleteAdditionalDiskShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	var actualStorageContainerName string
	var actualBlobName string

	var testSubject = &StepDeleteAdditionalDisk{
		delete: func(storageContainerName string, blobName string) error {
			actualStorageContainerName = storageContainerName
			actualBlobName = blobName
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteAdditionalDisk([]string{"http://storage.blob.core.windows.net/images/pkrvm_os.vhd"})
	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if actualStorageContainerName != "images" {
		t.Fatalf("Expected the storage container name to be 'images', but found '%s'.", actualStorageContainerName)
	}

	if actualBlobName != "pkrvm_os.vhd" {
		t.Fatalf("Expected the blob name to be 'pkrvm_os.vhd', but found '%s'.", actualBlobName)
	}
}

func TestStepDeleteAdditionalDiskShouldHandleComplexStorageContainerNames(t *testing.T) {
	var actualStorageContainerName string
	var actualBlobName string

	var testSubject = &StepDeleteAdditionalDisk{
		delete: func(storageContainerName string, blobName string) error {
			actualStorageContainerName = storageContainerName
			actualBlobName = blobName
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteAdditionalDisk([]string{"http://storage.blob.core.windows.net/abc/def/pkrvm_os.vhd"})
	testSubject.Run(context.Background(), stateBag)

	if actualStorageContainerName != "abc" {
		t.Fatalf("Expected the storage container name to be 'abc/def', but found '%s'.", actualStorageContainerName)
	}

	if actualBlobName != "def/pkrvm_os.vhd" {
		t.Fatalf("Expected the blob name to be 'pkrvm_os.vhd', but found '%s'.", actualBlobName)
	}
}

func TestStepDeleteAdditionalDiskShouldFailIfVHDNameCannotBeURLParsed(t *testing.T) {
	var testSubject = &StepDeleteAdditionalDisk{
		delete:        func(string, string) error { return nil },
		say:           func(message string) {},
		error:         func(e error) {},
		deleteManaged: func(context.Context, string, string) error { return nil },
	}

	// Invalid URL per https://golang.org/src/net/url/url_test.go
	stateBag := DeleteTestStateBagStepDeleteAdditionalDisk([]string{"http://[fe80::1%en0]/"})

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%v'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to not stateBag['%s'], but it was.", constants.Error)
	}
}
func TestStepDeleteAdditionalDiskShouldFailIfVHDNameIsTooShort(t *testing.T) {
	var testSubject = &StepDeleteAdditionalDisk{
		delete:        func(string, string) error { return nil },
		say:           func(message string) {},
		error:         func(e error) {},
		deleteManaged: func(context.Context, string, string) error { return nil },
	}

	stateBag := DeleteTestStateBagStepDeleteAdditionalDisk([]string{"storage.blob.core.windows.net/abc"})

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to not stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeleteAdditionalDiskShouldPassIfManagedDiskInTempResourceGroup(t *testing.T) {
	var testSubject = &StepDeleteAdditionalDisk{
		delete: func(string, string) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmAdditionalDiskVhds, []string{"subscriptions/123-456-789/resourceGroups/existingresourcegroup/providers/Microsoft.Compute/disks/osdisk"})
	stateBag.Put(constants.ArmIsManagedImage, true)
	stateBag.Put(constants.ArmIsExistingResourceGroup, false)
	stateBag.Put(constants.ArmResourceGroupName, "testgroup")

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeleteAdditionalDiskShouldFailIfManagedDiskInExistingResourceGroupFailsToDelete(t *testing.T) {
	var testSubject = &StepDeleteAdditionalDisk{
		delete:        func(string, string) error { return nil },
		say:           func(message string) {},
		error:         func(e error) {},
		deleteManaged: func(context.Context, string, string) error { return errors.New("UNIT TEST FAIL!") },
	}

	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmAdditionalDiskVhds, []string{"subscriptions/123-456-789/resourceGroups/existingresourcegroup/providers/Microsoft.Compute/disks/osdisk"})
	stateBag.Put(constants.ArmIsManagedImage, true)
	stateBag.Put(constants.ArmIsExistingResourceGroup, true)
	stateBag.Put(constants.ArmResourceGroupName, "testgroup")

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to not stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeleteAdditionalDiskShouldFailIfManagedDiskInExistingResourceGroupIsDeleted(t *testing.T) {
	var testSubject = &StepDeleteAdditionalDisk{
		delete:        func(string, string) error { return nil },
		say:           func(message string) {},
		error:         func(e error) {},
		deleteManaged: func(context.Context, string, string) error { return nil },
	}

	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmAdditionalDiskVhds, []string{"subscriptions/123-456-789/resourceGroups/existingresourcegroup/providers/Microsoft.Compute/disks/osdisk"})
	stateBag.Put(constants.ArmIsManagedImage, true)
	stateBag.Put(constants.ArmIsExistingResourceGroup, true)
	stateBag.Put(constants.ArmResourceGroupName, "testgroup")

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func DeleteTestStateBagStepDeleteAdditionalDisk(osDiskVhds []string) multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmAdditionalDiskVhds, osDiskVhds)
	stateBag.Put(constants.ArmIsManagedImage, false)
	stateBag.Put(constants.ArmIsExistingResourceGroup, false)
	stateBag.Put(constants.ArmResourceGroupName, "testgroup")

	return stateBag
}
