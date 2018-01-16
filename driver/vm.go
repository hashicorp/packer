package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"errors"
	"time"
	"fmt"
)

type VirtualMachine struct {
	vm     *object.VirtualMachine
	driver *Driver
}

type CloneConfig struct {
	Name         string
	Folder       string
	Host         string
	ResourcePool string
	Datastore    string
	LinkedClone  bool
}

type HardwareConfig struct {
	CPUs           int32
	CPUReservation int64
	CPULimit       int64
	RAM            int64
	RAMReservation int64
	RAMReserveAll  bool
	DiskSize       int64
	NestedHV       bool
}

func (d *Driver) NewVM(ref *types.ManagedObjectReference) *VirtualMachine {
	return &VirtualMachine{
		vm:     object.NewVirtualMachine(d.client.Client, *ref),
		driver: d,
	}
}

func (d *Driver) FindVM(name string) (*VirtualMachine, error) {
	vm, err := d.finder.VirtualMachine(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &VirtualMachine{
		vm:     vm,
		driver: d,
	}, nil
}

func (vm *VirtualMachine) Info(params ...string) (*mo.VirtualMachine, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.VirtualMachine
	err := vm.vm.Properties(vm.driver.ctx, vm.vm.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (template *VirtualMachine) Clone(config *CloneConfig) (*VirtualMachine, error) {
	folder, err := template.driver.FindFolder(config.Folder)
	if err != nil {
		return nil, err
	}

	var relocateSpec types.VirtualMachineRelocateSpec

	pool, err := template.driver.FindResourcePool(config.Host, config.ResourcePool)
	if err != nil {
		return nil, err
	}
	poolRef := pool.pool.Reference()
	relocateSpec.Pool = &poolRef

	if config.Datastore == "" {
		host, err := template.driver.FindHost(config.Host)
		if err != nil {
			return nil, err
		}

		info, err := host.Info("datastore")
		if err != nil {
			return nil, err
		}

		if len(info.Datastore) > 1 {
			return nil, fmt.Errorf("Target host has several datastores. Specify 'datastore' parameter explicitly")
		}

		ref := info.Datastore[0].Reference()
		relocateSpec.Datastore = &ref
	} else {
		ds, err := template.driver.FindDatastore(config.Datastore)
		if err != nil {
			return nil, err
		}

		ref := ds.ds.Reference()
		relocateSpec.Datastore = &ref
	}

	var cloneSpec types.VirtualMachineCloneSpec
	cloneSpec.Location = relocateSpec
	cloneSpec.PowerOn = false

	if config.LinkedClone == true {
		cloneSpec.Location.DiskMoveType = "createNewChildDiskBacking"

		tpl, err := template.Info("snapshot")
		if err != nil {
			return nil, err
		}
		if tpl.Snapshot == nil {
			err = errors.New("`linked_clone=true`, but template has no snapshots")
			return nil, err
		}
		cloneSpec.Snapshot = tpl.Snapshot.CurrentSnapshot
	}

	task, err := template.vm.Clone(template.driver.ctx, folder.folder, config.Name, cloneSpec)
	if err != nil {
		return nil, err
	}

	info, err := task.WaitForResult(template.driver.ctx, nil)
	if err != nil {
		return nil, err
	}

	ref := info.Result.(types.ManagedObjectReference)
	vm := template.driver.NewVM(&ref)
	return vm, nil
}

func (vm *VirtualMachine) Destroy() error {
	task, err := vm.vm.Destroy(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachine) Configure(config *HardwareConfig) error {
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
	confSpec.NestedHVEnabled = &config.NestedHV

	if config.DiskSize > 0 {
		devices, err := vm.vm.Device(vm.driver.ctx)
		if err != nil {
			return err
		}

		disk, err := findDisk(devices)
		if err != nil {
			return err
		}

		disk.CapacityInKB = config.DiskSize * 1024 * 1024 // Gb

		confSpec.DeviceChange = []types.BaseVirtualDeviceConfigSpec{
			&types.VirtualDeviceConfigSpec{
				Device:    disk,
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
			},
		}
	}

	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func findDisk(devices object.VirtualDeviceList) (*types.VirtualDisk, error) {
	var disks []*types.VirtualDisk
	for _, device := range devices {
		switch d := device.(type) {
		case *types.VirtualDisk:
			disks = append(disks, d)
		}
	}

	switch len(disks) {
	case 0:
		return nil, errors.New("VM has no disks")
	case 1:
		return disks[0], nil
	}
	return nil, errors.New("VM has multiple disks")
}

func (vm *VirtualMachine) PowerOn() error {
	task, err := vm.vm.PowerOn(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachine) WaitForIP() (string, error) {
	ip, err := vm.vm.WaitForIP(vm.driver.ctx)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (vm *VirtualMachine) PowerOff() error {
	state, err := vm.vm.PowerState(vm.driver.ctx)
	if err != nil {
		return err
	}

	if state == types.VirtualMachinePowerStatePoweredOff {
		return nil
	}

	task, err := vm.vm.PowerOff(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachine) StartShutdown() error {
	err := vm.vm.ShutdownGuest(vm.driver.ctx)
	return err
}

func (vm *VirtualMachine) WaitForShutdown(timeout time.Duration) error {
	shutdownTimer := time.After(timeout)
	for {
		powerState, err := vm.vm.PowerState(vm.driver.ctx)
		if err != nil {
			return err
		}
		if powerState == "poweredOff" {
			break
		}

		select {
		case <-shutdownTimer:
			err := errors.New("Timeout while waiting for machine to shut down.")
			return err
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func (vm *VirtualMachine) CreateSnapshot(name string) error {
	task, err := vm.vm.CreateSnapshot(vm.driver.ctx, name, "", false, false)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachine) ConvertToTemplate() error {
	err := vm.vm.MarkAsTemplate(vm.driver.ctx)
	return err
}
