package chroot

import (
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/stretchr/testify/assert"
)

func Test_MetadataReturnsVMResourceID(t *testing.T) {
	if !common.IsAzure() {
		t.Skipf("Not running on Azure, skipping live IMDS test")
	}
	mdc := NewMetadataClient()
	id, err := mdc.VMResourceID()
	assert.Nil(t, err)
	assert.NotEqual(t, id, "", "Expected VMResourceID to return non-empty string because we are running on Azure")

	vm, err := azure.ParseResourceID(id)
	assert.Nil(t, err, "%q is not parsable as an Azure resource id", id)
	t.Logf("VM: %+v", vm)
}
