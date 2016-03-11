// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"
	"testing"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common/constants"
)

func TestStepDeleteOSDiskShouldFailIfGetFails(t *testing.T) {
	var testSubject = &StepDeleteOSDisk{
		delete: func(string, string) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteOSDisk("http://storage.blob.core.windows.net/images/pkrvm_os.vhd")

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepDeleteOSDiskShouldPassIfGetPasses(t *testing.T) {
	var testSubject = &StepDeleteOSDisk{
		delete: func(string, string) error { return nil },
		say:    func(message string) {},
		error:  func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteOSDisk("http://storage.blob.core.windows.net/images/pkrvm_os.vhd")

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepDeleteOSDiskShouldTakeStepArgumentsFromStateBag(t *testing.T) {
	var actualStorageContainerName string
	var actualBlobName string

	var testSubject = &StepDeleteOSDisk{
		delete: func(storageContainerName string, blobName string) error {
			actualStorageContainerName = storageContainerName
			actualBlobName = blobName
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteOSDisk("http://storage.blob.core.windows.net/images/pkrvm_os.vhd")
	var result = testSubject.Run(stateBag)

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

func TestStepDeleteOSDiskShouldHandleComplexStorageContainerNames(t *testing.T) {
	var actualStorageContainerName string
	var actualBlobName string

	var testSubject = &StepDeleteOSDisk{
		delete: func(storageContainerName string, blobName string) error {
			actualStorageContainerName = storageContainerName
			actualBlobName = blobName
			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := DeleteTestStateBagStepDeleteOSDisk("http://storage.blob.core.windows.net/abc/def/pkrvm_os.vhd")
	testSubject.Run(stateBag)

	if actualStorageContainerName != "abc" {
		t.Fatalf("Expected the storage container name to be 'abc/def', but found '%s'.", actualStorageContainerName)
	}

	if actualBlobName != "def/pkrvm_os.vhd" {
		t.Fatalf("Expected the blob name to be 'pkrvm_os.vhd', but found '%s'.", actualBlobName)
	}
}

func DeleteTestStateBagStepDeleteOSDisk(osDiskVhd string) multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)
	stateBag.Put(constants.ArmOSDiskVhd, osDiskVhd)

	return stateBag
}
