package chroot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/packer/builder/azure/common/client"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

type VirtualMachinesClientAPI interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachine) (
		result compute.VirtualMachinesCreateOrUpdateFuture, err error)
	Get(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (
		result compute.VirtualMachine, err error)
}

type DiskAttacher interface {
	AttachDisk(ctx context.Context, disk string) (lun int32, err error)
	DetachDisk(ctx context.Context, disk string) (err error)
	WaitForDevice(ctx context.Context, i int32) (device string, err error)
}

func NewDiskAttacher(azureClient client.AzureClientSet) DiskAttacher {
	return diskAttacher{azureClient}
}

type diskAttacher struct {
	azcli client.AzureClientSet
}

func (da diskAttacher) WaitForDevice(ctx context.Context, i int32) (device string, err error) {
	path := fmt.Sprintf("/dev/disk/azure/scsi1/lun%d", i)

	for {
		l, err := os.Readlink(path)
		if err == nil {
			return filepath.Abs("/dev/disk/azure/scsi1/" + l)
		}
		if err != nil && err != os.ErrNotExist {
			return "", err
		}
		select {
		case <-time.After(100 * time.Millisecond):
			// continue
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func (da diskAttacher) DetachDisk(ctx context.Context, diskID string) error {
	currentDisks, err := da.getDisks(ctx)
	if err != nil {
		return err
	}

	// copy all disks to new array that not match diskID
	newDisks := []compute.DataDisk{}
	for _, disk := range currentDisks {
		if disk.ManagedDisk != nil &&
			!strings.EqualFold(to.String(disk.ManagedDisk.ID), diskID) {
			newDisks = append(newDisks, disk)
		}
	}
	if len(currentDisks) == len(newDisks) {
		return DiskNotFoundError
	}

	return da.setDisks(ctx, newDisks)
}

var DiskNotFoundError = errors.New("Disk not found")

func (da diskAttacher) AttachDisk(ctx context.Context, diskID string) (int32, error) {
	dataDisks, err := da.getDisks(ctx)
	if err != nil {
		return -1, err
	}

	// check to see if disk is already attached, remember lun if found
	var lun int32 = -1
	for _, disk := range dataDisks {
		if disk.ManagedDisk != nil &&
			strings.EqualFold(to.String(disk.ManagedDisk.ID), diskID) {
			// disk is already attached, just take this lun
			if disk.Lun != nil {
				lun = to.Int32(disk.Lun)
				break
			}
		}
	}

	if lun == -1 {
		// disk was not found on VM, go and actually attach it

	findFreeLun:
		for lun = 0; lun < 64; lun++ {
			for _, v := range dataDisks {
				if to.Int32(v.Lun) == lun {
					continue findFreeLun
				}
			}
			// no datadisk is using this lun
			break
		}

		// append new data disk to collection
		dataDisks = append(dataDisks, compute.DataDisk{
			CreateOption: compute.DiskCreateOptionTypesAttach,
			ManagedDisk: &compute.ManagedDiskParameters{
				ID: to.StringPtr(diskID),
			},
			Lun: to.Int32Ptr(lun),
		})

		// prepare resource object for update operation
		err = da.setDisks(ctx, dataDisks)
		if err != nil {
			return -1, err
		}
	}
	return lun, nil
}

func (da diskAttacher) getThisVM(ctx context.Context) (compute.VirtualMachine, error) {
	// getting resource info for this VM
	vm, err := da.azcli.MetadataClient().GetComputeInfo()
	if err != nil {
		return compute.VirtualMachine{}, err
	}

	// retrieve actual VM
	vmResource, err := da.azcli.VirtualMachinesClient().Get(ctx, vm.ResourceGroupName, vm.Name, "")
	if err != nil {
		return compute.VirtualMachine{}, err
	}
	if vmResource.StorageProfile == nil {
		return compute.VirtualMachine{}, errors.New("properties.storageProfile is not set on VM, this is unexpected")
	}

	return vmResource, nil
}

func (da diskAttacher) getDisks(ctx context.Context) ([]compute.DataDisk, error) {
	vmResource, err := da.getThisVM(ctx)
	if err != nil {
		return []compute.DataDisk{}, err
	}

	return *vmResource.StorageProfile.DataDisks, nil
}

func (da diskAttacher) setDisks(ctx context.Context, disks []compute.DataDisk) error {
	vmResource, err := da.getThisVM(ctx)
	if err != nil {
		return err
	}

	id, err := azure.ParseResourceID(to.String(vmResource.ID))
	if err != nil {
		return err
	}

	vmResource.StorageProfile.DataDisks = &disks
	vmResource.Resources = nil

	// update the VM resource, attaching disk
	f, err := da.azcli.VirtualMachinesClient().CreateOrUpdate(ctx, id.ResourceGroup, id.ResourceName, vmResource)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, da.azcli.PollClient())
	}
	return err
}
