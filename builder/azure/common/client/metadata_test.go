package client

import (
	"fmt"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/stretchr/testify/assert"
)

func Test_MetadataReturnsComputeInfo(t *testing.T) {
	if !IsAzure() {
		t.Skipf("Not running on Azure, skipping live IMDS test")
	}
	mdc := NewMetadataClient()
	info, err := mdc.GetComputeInfo()
	assert.Nil(t, err)

	vm, err := azure.ParseResourceID(fmt.Sprintf(
		"/subscriptions/%s"+
			"/resourceGroups/%s"+
			"/providers/Microsoft.Compute"+
			"/virtualMachines/%s",
		info.SubscriptionID,
		info.ResourceGroupName,
		info.Name))
	assert.Nil(t, err, "%q is not parsable as an Azure resource info", info)

	assert.Regexp(t, "^[0-9a-fA-F-]{36}$", vm.SubscriptionID)
	t.Logf("VM: %+v", vm)
}
