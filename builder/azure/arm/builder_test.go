package arm

import (
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
)

func TestStateBagShouldBePopulatedExpectedValues(t *testing.T) {
	var testSubject Builder
	_, _, err := testSubject.Prepare(getArmBuilderConfiguration(), getPackerConfiguration())
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

func TestStateBagShouldPoluateExpectedTags(t *testing.T) {
	var testSubject Builder

	expectedTags := map[string]string{
		"env":     "test",
		"builder": "packer",
	}
	armConfig := getArmBuilderConfiguration()
	armConfig["azure_tags"] = expectedTags

	_, _, err := testSubject.Prepare(armConfig, getPackerConfiguration())
	if err != nil {
		t.Fatalf("failed to prepare: %s", err)
	}

	tags, ok := testSubject.stateBag.Get(constants.ArmTags).(map[string]*string)
	if !ok {
		t.Errorf("Expected the builder's state bag to contain tags of type %T, but didn't.", testSubject.config.AzureTags)
	}

	if len(tags) != len(expectedTags) {
		t.Errorf("expect tags from state to be the same length as tags from config")
	}

	for k, v := range tags {
		if expectedTags[k] != *v {
			t.Errorf("expect tag value of %s to be %s, but got %s", k, expectedTags[k], *v)
		}
	}

}
