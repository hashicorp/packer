package main

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"context"
	"net/url"
	"fmt"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/mo"
	"errors"
)

type Driver struct {
	ctx        context.Context
	client     *govmomi.Client
	datacenter *object.Datacenter
	finder     *find.Finder
}

func NewDriverVSphere(config *ConnectConfig) (Driver, error) {
	ctx := context.TODO()

	vcenter_url, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return Driver{}, err
	}
	vcenter_url.User = url.UserPassword(config.Username, config.Password)
	client, err := govmomi.NewClient(ctx, vcenter_url, config.InsecureConnection)
	if err != nil {
		return Driver{}, err
	}

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return Driver{}, err
	}
	finder.SetDatacenter(datacenter)

	d := Driver{
		ctx:        ctx,
		client:     client,
		datacenter: datacenter,
		finder:     finder,
	}
	return d, nil
}

func (d *Driver) cloneVM(config *CloneConfig) (*object.VirtualMachine, error) {
	vmSrc, err := d.finder.VirtualMachine(d.ctx, config.Template)
	if err != nil {
		return nil, err
	}

	folder, err := d.finder.FolderOrDefault(d.ctx, fmt.Sprintf("/%v/vm/%v", d.datacenter.Name(), config.FolderName))
	if err != nil {
		return nil, err
	}

	pool, err := d.finder.ResourcePoolOrDefault(d.ctx, fmt.Sprintf("/%v/host/%v/Resources/%v", d.datacenter.Name(), config.Host, config.ResourcePool))
	if err != nil {
		return nil, err
	}
	poolRef := pool.Reference()

	var datastore *object.Datastore
	if config.Datastore != "" {
		datastore, err = d.finder.Datastore(d.ctx, config.Datastore)
		if err != nil {
			return nil, err
		}
	}

	// Creating specs for cloning
	relocateSpec := types.VirtualMachineRelocateSpec{
		Pool: &(poolRef),
	}
	if datastore != nil {
		datastoreRef := datastore.Reference()
		relocateSpec.Datastore = &datastoreRef
	}
	if config.LinkedClone == true {
		relocateSpec.DiskMoveType = "createNewChildDiskBacking"
	}

	cloneSpec := types.VirtualMachineCloneSpec{
		Location: relocateSpec,
		PowerOn:  false,
	}
	if config.LinkedClone == true {
		var vmImage mo.VirtualMachine
		err = vmSrc.Properties(d.ctx, vmSrc.Reference(), []string{"snapshot"}, &vmImage)
		if err != nil {
			err = fmt.Errorf("Error reading base VM properties: %s", err)
			return nil, err
		}
		if vmImage.Snapshot == nil {
			err = errors.New("`linked_clone=true`, but image VM has no snapshots")
			return nil, err
		}
		cloneSpec.Snapshot = vmImage.Snapshot.CurrentSnapshot
	}

	// Cloning itself
	task, err := vmSrc.Clone(d.ctx, folder, config.VMName, cloneSpec)
	if err != nil {
		return nil, err
	}

	info, err := task.WaitForResult(d.ctx, nil)
	if err != nil {
		return nil, err
	}

	vm := object.NewVirtualMachine(vmSrc.Client(), info.Result.(types.ManagedObjectReference))
	return vm, nil
}

func (d *Driver) destroyVM(vm *object.VirtualMachine) error {
	task, err := vm.Destroy(d.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) configureVM(vm *object.VirtualMachine, config *HardwareConfig) error {
	var confSpec types.VirtualMachineConfigSpec
	confSpec.NumCPUs = config.CPUs
	confSpec.MemoryMB = config.RAM

	var cpuSpec types.ResourceAllocationInfo
	cpuSpec.Reservation = config.CPUReservation
	cpuSpec.Limit = config.CPULimit
	confSpec.CpuAllocation = &cpuSpec

	var ramSpec types.ResourceAllocationInfo
	ramSpec.Reservation = config.RAMReservation
	confSpec.MemoryAllocation = &ramSpec

	confSpec.MemoryReservationLockedToMax = &config.RAMReserveAll

	task, err := vm.Reconfigure(d.ctx, confSpec)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
