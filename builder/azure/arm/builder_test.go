package arm

import (
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
)

func TestStateBagShouldBePopulatedExpectedValues(t *testing.T) {
	var testSubject = &Builder{}
	_, err := testSubject.Prepare(getArmBuilderConfiguration(), getPackerConfiguration())
	if err != nil {
		t.Fatalf("failed to prepare: %s", err)
	}

	var expectedStateBagKeys = []string{
		constants.AuthorizedKey,

		constants.ArmTags,
		constants.ArmComputeName,
		constants.ArmDeploymentName,
		constants.ArmNicName,
		constants.ArmResourceGroupName,
		constants.ArmStorageAccountName,
		constants.ArmVirtualMachineCaptureParameters,
		constants.ArmPublicIPAddressName,
		constants.ArmAsyncResourceGroupDelete,
	}

	for _, v := range expectedStateBagKeys {
		if _, ok := testSubject.stateBag.GetOk(v); ok == false {
			t.Errorf("Expected the builder's state bag to contain '%s', but it did not.", v)
		}
	}
}
