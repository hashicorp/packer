package chroot

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
)

type DiskAttacher interface {
	AttachDisk(ctx context.Context, disk string) (lun int32, err error)
	WaitForDevice(ctx context.Context, i int32) (device string, err error)
	DetachDisk(ctx context.Context, disk string) (err error)
	WaitForDetach(ctx context.Context, diskID string) error
}

var NewDiskAttacher = func(azureClient client.AzureClientSet) DiskAttacher {
	return &diskAttacher{
		azcli: azureClient,
	}
}

type diskAttacher struct {
	azcli client.AzureClientSet

	vm *client.ComputeInfo // store info about this VM so that we don't have to ask metadata service on every call
}

var DiskNotFoundError = errors.New("Disk not found")

func (da *diskAttacher) AttachDisk(ctx context.Context, diskID string) (int32, error) {
	dataDisks, err := da.getDisks(ctx)
	if err != nil {
		return -1, err
	}

	// check to see if disk is already attached, remember lun if found
	if disk := findDiskInList(dataDisks, diskID); disk != nil {
		// disk is already attached, just take this lun
		if disk.Lun == nil {
			return -1, errors.New("disk is attached, but lun was not set in VM model (possibly an error in the Azure APIs)")
		}
		return to.Int32(disk.Lun), nil
	}

	// disk was not found on VM, go and actually attach it

	var lun int32 = -1
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

	return lun, nil
}

func (da *diskAttacher) DetachDisk(ctx context.Context, diskID string) error {
	log.Println("Fetching list of disks currently attached to VM")
	currentDisks, err := da.getDisks(ctx)
	if err != nil {
		return err
	}

	log.Printf("Removing %q from list of disks currently attached to VM", diskID)
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

	log.Println("Updating new list of disks attached to VM")
	err = da.setDisks(ctx, newDisks)
	if err != nil {
		return err
	}

	return nil
}

func (da *diskAttacher) WaitForDetach(ctx context.Context, diskID string) error {
	for { // loop until disk is not attached, timeout or error
		list, err := da.getDisks(ctx)
		if err != nil {
			return err
		}
		if findDiskInList(list, diskID) == nil {
			log.Println("Disk is no longer in VM model, assuming detached")
			return nil
		}

		select {
		case <-time.After(time.Second): //continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (da *diskAttacher) getThisVM(ctx context.Context) (compute.VirtualMachine, error) {
	// getting resource info for this VM
	if da.vm == nil {
		vm, err := da.azcli.MetadataClient().GetComputeInfo()
		if err != nil {
			return compute.VirtualMachine{}, err
		}
		da.vm = vm
	}

	// retrieve actual VM
	vmResource, err := da.azcli.VirtualMachinesClient().Get(ctx, da.vm.ResourceGroupName, da.vm.Name, "")
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

	// update the VM resource, attach disk
	_, err = da.azcli.VirtualMachinesClient().CreateOrUpdate(ctx, id.ResourceGroup, id.ResourceName, vmResource)

	return err
}

func findDiskInList(list []compute.DataDisk, diskID string) *compute.DataDisk {
	for _, disk := range list {
		if disk.ManagedDisk != nil &&
			strings.EqualFold(to.String(disk.ManagedDisk.ID), diskID) {
			return &disk
		}
	}
	return nil
}
