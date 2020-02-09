package chroot

import (
	"context"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testvm   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testGroup/Microsoft.Compute/virtualMachines/testVM"
	testdisk = "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/testGroup2/Microsoft.Compute/disks/testDisk"
)

// Tests assume current machine is capable of running chroot builder (i.e. an Azure VM)

func Test_DiskAttacherAttachesDiskToVM(t *testing.T) {
	azcli, err := client.GetTestClientSet(t) // integration test
	require.Nil(t, err)
	da := NewDiskAttacher(azcli)
	testDiskName := t.Name()

	vm, err := azcli.MetadataClient().GetComputeInfo()
	require.Nil(t, err, "Test needs to run on an Azure VM, unable to retrieve VM information")
	t.Log("Creating new disk '", testDiskName, "' in ", vm.ResourceGroupName)

	disk, err := azcli.DisksClient().Get(context.TODO(), vm.ResourceGroupName, testDiskName)
	if err == nil {
		t.Log("Disk already exists")
		if disk.DiskState == compute.Attached {
			t.Log("Disk is attached, assuming to this machine, trying to detach")
			err = da.DetachDisk(context.TODO(), to.String(disk.ID))
			require.Nil(t, err)
		}
		t.Log("Deleting disk")
		result, err := azcli.DisksClient().Delete(context.TODO(), vm.ResourceGroupName, testDiskName)
		require.Nil(t, err)
		err = result.WaitForCompletionRef(context.TODO(), azcli.PollClient())
		require.Nil(t, err)
	}

	t.Log("Creating disk")
	r, err := azcli.DisksClient().CreateOrUpdate(context.TODO(), vm.ResourceGroupName, testDiskName, compute.Disk{
		Location: to.StringPtr(vm.Location),
		Sku: &compute.DiskSku{
			Name: compute.StandardLRS,
		},
		DiskProperties: &compute.DiskProperties{
			DiskSizeGB:   to.Int32Ptr(30),
			CreationData: &compute.CreationData{CreateOption: compute.Empty},
		},
	})
	require.Nil(t, err)
	err = r.WaitForCompletionRef(context.TODO(), azcli.PollClient())
	require.Nil(t, err)

	t.Log("Retrieving disk properties")
	d, err := azcli.DisksClient().Get(context.TODO(), vm.ResourceGroupName, testDiskName)
	require.Nil(t, err)
	assert.NotNil(t, d)

	t.Log("Attaching disk")
	lun, err := da.AttachDisk(context.TODO(), to.String(d.ID))
	assert.Nil(t, err)

	t.Log("Waiting for device")
	dev, err := da.WaitForDevice(context.TODO(), lun)
	assert.Nil(t, err)

	t.Log("Device path:", dev)

	t.Log("Detaching disk")
	err = da.DetachDisk(context.TODO(), to.String(d.ID))
	require.Nil(t, err)

	t.Log("Deleting disk")
	result, err := azcli.DisksClient().Delete(context.TODO(), vm.ResourceGroupName, testDiskName)
	if err == nil {
		err = result.WaitForCompletionRef(context.TODO(), azcli.PollClient())
	}
	require.Nil(t, err)
}
