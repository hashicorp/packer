// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"testing"
)

func TestStateBagShouldBePopulatedExpectedValues(t *testing.T) {
	var testSubject = &Builder{}
	_, err := testSubject.Prepare(getArmBuilderConfiguration(), getPackerConfiguration())
	if err != nil {
		t.Fatalf("failed to prepare: %s", err)
	}

	var expectedStateBagKeys = []string{
		constants.AuthorizedKey,
		constants.PrivateKey,

		constants.ArmComputeName,
		constants.ArmDeploymentName,
		constants.ArmLocation,
		constants.ArmResourceGroupName,
		constants.ArmStorageAccountName,
		constants.ArmTemplateParameters,
		constants.ArmVirtualMachineCaptureParameters,
		constants.ArmPublicIPAddressName,
	}

	for _, v := range expectedStateBagKeys {
		if _, ok := testSubject.stateBag.GetOk(v); ok == false {
			t.Errorf("Expected the builder's state bag to contain '%s', but it did not.", v)
		}
	}
}
