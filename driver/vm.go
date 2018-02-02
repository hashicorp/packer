package driver

import (
	"errors"
	"fmt"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"time"
	"strings"
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

type CreateConfig struct {
	HardwareConfig

	DiskThinProvisioned bool
	DiskControllerType  string // example: "scsi", "pvscsi"

	Annotation   string
	Name         string
	Folder       string
	Host         string
	ResourcePool string
	Datastore    string
	GuestOS      string // example: otherGuest
	Network      string // "" for default network
	NetworkCard  string // example: vmxnet3
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

func (d *Driver) CreateVM(config *CreateConfig) (*VirtualMachine, error) {
	createSpec := config.toConfigSpec()

	folder, err := d.FindFolder(config.Folder)
	if err != nil {
		return nil, err
	}

	resourcePool, err := d.FindResourcePool(config.Host, config.ResourcePool)
	if err != nil {
		return nil, err
	}

	host, err := d.FindHost(config.Host)
	if err != nil {
		return nil, err
	}

	datastore, err := d.FindDatastore(config.Datastore)
	if err != nil {
		return nil, err
	}

	devices := object.VirtualDeviceList{}

	devices, err = addIDE(devices)
	if err != nil {
		return nil, err
	}
	devices, err = addDisk(d, devices, config)
	if err != nil {
		return nil, err
	}
	devices, err = addNetwork(d, devices, config)
	if err != nil {
		return nil, err
	}

	createSpec.DeviceChange, err = devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return nil, err
	}

	createSpec.Files = &types.VirtualMachineFileInfo{
		VmPathName: fmt.Sprintf("[%s]", datastore.Name()),
	}

	task, err := folder.folder.CreateVM(d.ctx, createSpec, resourcePool.pool, host.host)
	if err != nil {
		return nil, err
	}
	taskInfo, err := task.WaitForResult(d.ctx, nil)
	if err != nil {
		return nil, err
	}

	vmRef := taskInfo.Result.(types.ManagedObjectReference)

	return d.NewVM(&vmRef), nil
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

func (vm *VirtualMachine) Devices() (object.VirtualDeviceList, error) {
	vmInfo, err := vm.Info("config.hardware.device")
	if err != nil {
		return nil, err
	}

	return vmInfo.Config.Hardware.Device, nil
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

	datastore, err := template.driver.FindDatastore(config.Datastore)
	if err != nil {
		return nil, err
	}
	datastoreRef := datastore.ds.Reference()
	relocateSpec.Datastore = &datastoreRef

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

	vmRef := info.Result.(types.ManagedObjectReference)
	vm := template.driver.NewVM(&vmRef)
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
	confSpec := config.toConfigSpec()

	if config.DiskSize > 0 {
		devices, err := vm.vm.Device(vm.driver.ctx)
		if err != nil {
			return err
		}

		disk, err := findDisk(devices)
		if err != nil {
			return err
		}

		disk.CapacityInKB = convertGiBToKiB(config.DiskSize)

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
	return vm.vm.WaitForIP(vm.driver.ctx)
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
	return vm.vm.MarkAsTemplate(vm.driver.ctx)
}

func (vm *VirtualMachine) GetDir() (string, error) {
	vmInfo, err := vm.Info("name", "layoutEx.file")
	if err != nil {
		return "", err
	}

	vmxName := fmt.Sprintf("/%s.vmx", vmInfo.Name)
	for _, file := range vmInfo.LayoutEx.File {
		if strings.HasSuffix(file.Name, vmxName) {
			return RemoveDatastorePrefix(file.Name[:len(file.Name)-len(vmxName)]), nil
		}
	}
	return "", fmt.Errorf("cannot find '%s'", vmxName)
}

func (config HardwareConfig) toConfigSpec() types.VirtualMachineConfigSpec {
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

	return confSpec
}

func (config CreateConfig) toConfigSpec() types.VirtualMachineConfigSpec {
	confSpec := config.HardwareConfig.toConfigSpec()
	confSpec.Name = config.Name
	confSpec.Annotation = config.Annotation
	confSpec.GuestId = config.GuestOS
	return confSpec
}

func addDisk(_ *Driver, devices object.VirtualDeviceList, config *CreateConfig) (object.VirtualDeviceList, error) {
	device, err := devices.CreateSCSIController(config.DiskControllerType)
	if err != nil {
		return nil, err
	}
	devices = append(devices, device)
	controller, err := devices.FindDiskController(devices.Name(device))
	if err != nil {
		return nil, err
	}

	disk := &types.VirtualDisk{
		VirtualDevice: types.VirtualDevice{
			Key: devices.NewKey(),
			Backing: &types.VirtualDiskFlatVer2BackingInfo{
				DiskMode:        string(types.VirtualDiskModePersistent),
				ThinProvisioned: types.NewBool(config.DiskThinProvisioned),
			},
		},
		CapacityInKB: convertGiBToKiB(config.DiskSize),
	}

	devices.AssignController(disk, controller)
	devices = append(devices, disk)

	return devices, nil
}

func addNetwork(d *Driver, devices object.VirtualDeviceList, config *CreateConfig) (object.VirtualDeviceList, error) {
	network, err := d.finder.NetworkOrDefault(d.ctx, config.Network)
	if err != nil {
		return nil, err
	}

	backing, err := network.EthernetCardBackingInfo(d.ctx)
	if err != nil {
		return nil, err
	}

	device, err := object.EthernetCardTypes().CreateEthernetCard(config.NetworkCard, backing)
	if err != nil {
		return nil, err
	}

	return append(devices, device), nil
}

func addIDE(devices object.VirtualDeviceList) (object.VirtualDeviceList, error) {
	ideDevice, err := devices.CreateIDEController()
	if err != nil {
		return nil, err
	}
	devices = append(devices, ideDevice)

	return devices, nil
}

func (vm *VirtualMachine) AddCdrom(isoPath string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}
	ide, err := devices.FindIDEController("")
	if err != nil {
		return err
	}

	cdrom, err := devices.CreateCdrom(ide)
	if err != nil {
		return err
	}

	if isoPath != "" {
		cdrom = devices.InsertIso(cdrom, isoPath)
	}

	return vm.addDevice(cdrom)
}

func (vm *VirtualMachine) AddFloppy(imgPath string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	floppy, err := devices.CreateFloppy()
	if err != nil {
		return err
	}

	if imgPath != "" {
		floppy = devices.InsertImg(floppy, imgPath)
	}

	return vm.addDevice(floppy)
}

func (vm *VirtualMachine) SetBootOrder(order []string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	bootOptions := types.VirtualMachineBootOptions{
		BootOrder: devices.BootOrder(order),
	}

	return vm.vm.SetBootOptions(vm.driver.ctx, &bootOptions)
}

func (vm *VirtualMachine) RemoveDevice(keepFiles bool, device ...types.BaseVirtualDevice) error {
	return vm.vm.RemoveDevice(vm.driver.ctx, keepFiles, device...)
}

func (vm *VirtualMachine) addDevice(device types.BaseVirtualDevice) error {
	newDevices := object.VirtualDeviceList{device}
	confSpec := types.VirtualMachineConfigSpec{}
	var err error
	confSpec.DeviceChange, err = newDevices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return err
	}

	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func convertGiBToKiB(gib int64) int64 {
	return gib * 1024 * 1024
}
