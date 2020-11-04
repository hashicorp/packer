package driver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VirtualMachine interface {
	Info(params ...string) (*mo.VirtualMachine, error)
	Devices() (object.VirtualDeviceList, error)
	FloppyDevices() (object.VirtualDeviceList, error)
	Clone(ctx context.Context, config *CloneConfig) (VirtualMachine, error)
	updateVAppConfig(ctx context.Context, newProps map[string]string) (*types.VmConfigSpec, error)
	AddPublicKeys(ctx context.Context, publicKeys string) error
	Properties(ctx context.Context) (*mo.VirtualMachine, error)
	Destroy() error
	Configure(config *HardwareConfig) error
	Customize(spec types.CustomizationSpec) error
	ResizeDisk(diskSize int64) error
	WaitForIP(ctx context.Context, ipNet *net.IPNet) (string, error)
	PowerOn() error
	PowerOff() error
	IsPoweredOff() (bool, error)
	StartShutdown() error
	WaitForShutdown(ctx context.Context, timeout time.Duration) error
	CreateSnapshot(name string) error
	ConvertToTemplate() error
	ImportOvfToContentLibrary(ovf vcenter.OVF) error
	ImportToContentLibrary(template vcenter.Template) error
	GetDir() (string, error)
	AddFloppy(imgPath string) error
	SetBootOrder(order []string) error
	RemoveDevice(keepFiles bool, device ...types.BaseVirtualDevice) error
	addDevice(device types.BaseVirtualDevice) error
	AddConfigParams(params map[string]string, info *types.ToolsConfigInfo) error
	Export() (*nfc.Lease, error)
	CreateDescriptor(m *ovf.Manager, cdp types.OvfCreateDescriptorParams) (*types.OvfCreateDescriptorResult, error)
	NewOvfManager() *ovf.Manager
	GetOvfExportOptions(m *ovf.Manager) ([]types.OvfOptionInfo, error)

	AddCdrom(controllerType string, datastoreIsoPath string) error
	CreateCdrom(c *types.VirtualController) (*types.VirtualCdrom, error)
	RemoveCdroms() error
	EjectCdroms() error
	AddSATAController() error
	FindSATAController() (*types.VirtualAHCIController, error)
}

type VirtualMachineDriver struct {
	vm     *object.VirtualMachine
	driver *VCenterDriver
}

type CloneConfig struct {
	Name           string
	Folder         string
	Cluster        string
	Host           string
	ResourcePool   string
	Datastore      string
	LinkedClone    bool
	Network        string
	MacAddress     string
	Annotation     string
	VAppProperties map[string]string
}

type HardwareConfig struct {
	CPUs                int32
	CpuCores            int32
	CPUReservation      int64
	CPULimit            int64
	RAM                 int64
	RAMReservation      int64
	RAMReserveAll       bool
	NestedHV            bool
	CpuHotAddEnabled    bool
	MemoryHotAddEnabled bool
	VideoRAM            int64
	VGPUProfile         string
	Firmware            string
	ForceBIOSSetup      bool
}

type NIC struct {
	Network     string // "" for default network
	NetworkCard string // example: vmxnet3
	MacAddress  string // set mac if want specific address
	Passthrough *bool  // direct path i/o
}

type CreateConfig struct {
	DiskControllerType []string // example: "scsi", "pvscsi", "nvme", "lsilogic"

	Annotation    string
	Name          string
	Folder        string
	Cluster       string
	Host          string
	ResourcePool  string
	Datastore     string
	GuestOS       string // example: otherGuest
	NICs          []NIC
	USBController []string
	Version       uint // example: 10
	Storage       []Disk
}

type Disk struct {
	DiskSize            int64
	DiskEagerlyScrub    bool
	DiskThinProvisioned bool
	ControllerIndex     int
}

func (d *VCenterDriver) NewVM(ref *types.ManagedObjectReference) VirtualMachine {
	return &VirtualMachineDriver{
		vm:     object.NewVirtualMachine(d.client.Client, *ref),
		driver: d,
	}
}

func (d *VCenterDriver) FindVM(name string) (VirtualMachine, error) {
	vm, err := d.finder.VirtualMachine(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &VirtualMachineDriver{
		vm:     vm,
		driver: d,
	}, nil
}

func (d *VCenterDriver) PreCleanVM(ui packer.Ui, vmPath string, force bool) error {
	vm, err := d.FindVM(vmPath)
	if err != nil {
		if _, ok := err.(*find.NotFoundError); !ok {
			return fmt.Errorf("error looking up old vm: %v", err)
		}
	}
	if force && vm != nil {
		ui.Say(fmt.Sprintf("the vm/template %s already exists, but deleting it due to -force flag", vmPath))

		// power off just in case it is still on
		vm.PowerOff()

		err := vm.Destroy()
		if err != nil {
			return fmt.Errorf("error destroying %s: %v", vmPath, err)
		}
	}
	if !force && vm != nil {
		return fmt.Errorf("%s already exists, you can use -force flag to destroy it: %v", vmPath, err)
	}

	return nil
}

func (d *VCenterDriver) CreateVM(config *CreateConfig) (VirtualMachine, error) {
	createSpec := types.VirtualMachineConfigSpec{
		Name:       config.Name,
		Annotation: config.Annotation,
		GuestId:    config.GuestOS,
	}
	if config.Version != 0 {
		createSpec.Version = fmt.Sprintf("%s%d", "vmx-", config.Version)
	}

	folder, err := d.FindFolder(config.Folder)
	if err != nil {
		return nil, err
	}

	resourcePool, err := d.FindResourcePool(config.Cluster, config.Host, config.ResourcePool)
	if err != nil {
		return nil, err
	}

	var host *object.HostSystem
	if config.Cluster != "" && config.Host != "" {
		h, err := d.FindHost(config.Host)
		if err != nil {
			return nil, err
		}
		host = h.host
	}

	datastore, err := d.FindDatastore(config.Datastore, config.Host)
	if err != nil {
		return nil, err
	}

	devices := object.VirtualDeviceList{}

	devices, err = addDisk(d, devices, config)
	if err != nil {
		return nil, err
	}
	devices, err = addNetwork(d, devices, config)
	if err != nil {
		return nil, err
	}

	t := true
	for _, usbType := range config.USBController {
		var usb types.BaseVirtualDevice
		switch usbType {
		// handle "true" and "1" for backwards compatibility
		case "usb", "true", "1":
			usb = &types.VirtualUSBController{
				EhciEnabled: &t,
			}
		case "xhci":
			usb = new(types.VirtualUSBXHCIController)
		default:
			continue
		}

		devices = append(devices, usb)
	}

	createSpec.DeviceChange, err = devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return nil, err
	}

	createSpec.Files = &types.VirtualMachineFileInfo{
		VmPathName: fmt.Sprintf("[%s]", datastore.Name()),
	}

	task, err := folder.folder.CreateVM(d.ctx, createSpec, resourcePool.pool, host)
	if err != nil {
		return nil, err
	}
	taskInfo, err := task.WaitForResult(d.ctx, nil)
	if err != nil {
		return nil, err
	}

	vmRef, ok := taskInfo.Result.(types.ManagedObjectReference)
	if !ok {
		return nil, fmt.Errorf("something went wrong when creating the VM")
	}

	return d.NewVM(&vmRef), nil
}

func (vm *VirtualMachineDriver) Info(params ...string) (*mo.VirtualMachine, error) {
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

func (vm *VirtualMachineDriver) Devices() (object.VirtualDeviceList, error) {
	vmInfo, err := vm.Info("config.hardware.device")
	if err != nil {
		return nil, err
	}

	return vmInfo.Config.Hardware.Device, nil
}

func (vm *VirtualMachineDriver) FloppyDevices() (object.VirtualDeviceList, error) {
	device, err := vm.Devices()
	if err != nil {
		return device, err
	}
	floppies := device.SelectByType((*types.VirtualFloppy)(nil))
	return floppies, nil
}

func (vm *VirtualMachineDriver) Clone(ctx context.Context, config *CloneConfig) (VirtualMachine, error) {
	folder, err := vm.driver.FindFolder(config.Folder)
	if err != nil {
		return nil, fmt.Errorf("Error finding filder: %s", err)
	}

	var relocateSpec types.VirtualMachineRelocateSpec

	pool, err := vm.driver.FindResourcePool(config.Cluster, config.Host, config.ResourcePool)
	if err != nil {
		return nil, fmt.Errorf("Error finding resource pool: %s", err)
	}
	poolRef := pool.pool.Reference()
	relocateSpec.Pool = &poolRef

	datastore, err := vm.driver.FindDatastore(config.Datastore, config.Host)
	if err != nil {
		return nil, fmt.Errorf("Error finding datastore: %s", err)
	}
	datastoreRef := datastore.Reference()
	relocateSpec.Datastore = &datastoreRef

	var cloneSpec types.VirtualMachineCloneSpec
	cloneSpec.Location = relocateSpec
	cloneSpec.PowerOn = false

	if config.LinkedClone == true {
		cloneSpec.Location.DiskMoveType = "createNewChildDiskBacking"

		tpl, err := vm.Info("snapshot")
		if err != nil {
			return nil, fmt.Errorf("Error getting snapshot info for vm: %s", err)
		}
		if tpl.Snapshot == nil {
			err = errors.New("`linked_clone=true`, but template has no snapshots")
			return nil, err
		}
		cloneSpec.Snapshot = tpl.Snapshot.CurrentSnapshot
	}

	var configSpec types.VirtualMachineConfigSpec
	cloneSpec.Config = &configSpec

	if config.Annotation != "" {
		configSpec.Annotation = config.Annotation
	}

	if config.Network != "" {
		net, err := vm.driver.FindNetwork(config.Network)
		if err != nil {
			return nil, fmt.Errorf("Error finding network: %s", err)
		}
		backing, err := net.network.EthernetCardBackingInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("Error finding ethernet card backing info: %s", err)
		}

		devices, err := vm.vm.Device(ctx)
		if err != nil {
			return nil, fmt.Errorf("Error finding vm devices: %s", err)
		}

		adapter, err := findNetworkAdapter(devices)
		if err != nil {
			return nil, fmt.Errorf("Error finding network adapter: %s", err)
		}

		current := adapter.GetVirtualEthernetCard()
		current.Backing = backing
		current.MacAddress = config.MacAddress

		config := &types.VirtualDeviceConfigSpec{
			Device:    adapter.(types.BaseVirtualDevice),
			Operation: types.VirtualDeviceConfigSpecOperationEdit,
		}

		configSpec.DeviceChange = append(configSpec.DeviceChange, config)
	}

	vAppConfig, err := vm.updateVAppConfig(ctx, config.VAppProperties)
	if err != nil {
		return nil, fmt.Errorf("Error updating VAppConfig: %s", err)
	}
	configSpec.VAppConfig = vAppConfig

	task, err := vm.vm.Clone(vm.driver.ctx, folder.folder, config.Name, cloneSpec)
	if err != nil {
		return nil, fmt.Errorf("Error calling vm.vm.Clone task: %s", err)
	}

	info, err := task.WaitForResult(ctx, nil)
	if err != nil {
		if ctx.Err() == context.Canceled {
			err = task.Cancel(context.TODO())
			return nil, err
		}

		return nil, fmt.Errorf("Error waiting for vm Clone to complete: %s", err)
	}

	vmRef, ok := info.Result.(types.ManagedObjectReference)
	if !ok {
		return nil, fmt.Errorf("something went wrong when cloning the VM")
	}

	created := vm.driver.NewVM(&vmRef)
	return created, nil
}

func (vm *VirtualMachineDriver) updateVAppConfig(ctx context.Context, newProps map[string]string) (*types.VmConfigSpec, error) {
	if len(newProps) == 0 {
		return nil, nil
	}

	vProps, _ := vm.Properties(ctx)
	if vProps.Config.VAppConfig == nil {
		return nil, fmt.Errorf("this VM lacks a vApp configuration and cannot have vApp properties set on it")
	}

	allProperties := vProps.Config.VAppConfig.GetVmConfigInfo().Property

	var props []types.VAppPropertySpec
	for _, p := range allProperties {
		userValue, setByUser := newProps[p.Id]
		if !setByUser {
			continue
		}

		if *p.UserConfigurable == false {
			return nil, fmt.Errorf("vApp property with userConfigurable=false specified in vapp.properties: %+v", reflect.ValueOf(newProps).MapKeys())
		}

		prop := types.VAppPropertySpec{
			ArrayUpdateSpec: types.ArrayUpdateSpec{
				Operation: types.ArrayUpdateOperationEdit,
			},
			Info: &types.VAppPropertyInfo{
				Key:              p.Key,
				Id:               p.Id,
				Value:            userValue,
				UserConfigurable: p.UserConfigurable,
			},
		}
		props = append(props, prop)

		delete(newProps, p.Id)
	}

	if len(newProps) > 0 {
		return nil, fmt.Errorf("unsupported vApp properties in vapp.properties: %+v", reflect.ValueOf(newProps).MapKeys())
	}

	return &types.VmConfigSpec{
		Property: props,
	}, nil
}

func (vm *VirtualMachineDriver) AddPublicKeys(ctx context.Context, publicKeys string) error {
	newProps := map[string]string{"public-keys": publicKeys}
	config, err := vm.updateVAppConfig(ctx, newProps)
	if err != nil {
		return fmt.Errorf("not possible to save temporary public key: %s", err.Error())
	}

	confSpec := types.VirtualMachineConfigSpec{VAppConfig: config}
	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachineDriver) Properties(ctx context.Context) (*mo.VirtualMachine, error) {
	log.Printf("fetching properties for VM %q", vm.vm.InventoryPath)
	var props mo.VirtualMachine
	if err := vm.vm.Properties(ctx, vm.vm.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

func (vm *VirtualMachineDriver) Destroy() error {
	task, err := vm.vm.Destroy(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachineDriver) Configure(config *HardwareConfig) error {
	var confSpec types.VirtualMachineConfigSpec
	confSpec.NumCPUs = config.CPUs
	confSpec.NumCoresPerSocket = config.CpuCores
	confSpec.MemoryMB = config.RAM

	var cpuSpec types.ResourceAllocationInfo
	cpuSpec.Reservation = &config.CPUReservation
	if config.CPULimit != 0 {
		cpuSpec.Limit = &config.CPULimit
	}
	confSpec.CpuAllocation = &cpuSpec

	var ramSpec types.ResourceAllocationInfo
	ramSpec.Reservation = &config.RAMReservation
	confSpec.MemoryAllocation = &ramSpec

	confSpec.MemoryReservationLockedToMax = &config.RAMReserveAll
	confSpec.NestedHVEnabled = &config.NestedHV

	confSpec.CpuHotAddEnabled = &config.CpuHotAddEnabled
	confSpec.MemoryHotAddEnabled = &config.MemoryHotAddEnabled

	if config.VideoRAM != 0 {
		devices, err := vm.vm.Device(vm.driver.ctx)
		if err != nil {
			return err
		}
		l := devices.SelectByType((*types.VirtualMachineVideoCard)(nil))
		if len(l) != 1 {
			return err
		}
		card := l[0].(*types.VirtualMachineVideoCard)

		card.VideoRamSizeInKB = config.VideoRAM

		spec := &types.VirtualDeviceConfigSpec{
			Device:    card,
			Operation: types.VirtualDeviceConfigSpecOperationEdit,
		}
		confSpec.DeviceChange = append(confSpec.DeviceChange, spec)
	}
	if config.VGPUProfile != "" {
		devices, err := vm.vm.Device(vm.driver.ctx)
		if err != nil {
			return err
		}

		pciDevices := devices.SelectByType((*types.VirtualPCIPassthrough)(nil))
		vGPUDevices := pciDevices.SelectByBackingInfo((*types.VirtualPCIPassthroughVmiopBackingInfo)(nil))
		var operation types.VirtualDeviceConfigSpecOperation
		if len(vGPUDevices) > 1 {
			return err
		} else if len(pciDevices) == 1 {
			operation = types.VirtualDeviceConfigSpecOperationEdit
		} else if len(pciDevices) == 0 {
			operation = types.VirtualDeviceConfigSpecOperationAdd
		}

		vGPUProfile := newVGPUProfile(config.VGPUProfile)
		spec := &types.VirtualDeviceConfigSpec{
			Device:    &vGPUProfile,
			Operation: operation,
		}
		log.Printf("Adding vGPU device with profile '%s'", config.VGPUProfile)
		confSpec.DeviceChange = append(confSpec.DeviceChange, spec)
	}

	efiSecureBootEnabled := false
	firmware := config.Firmware

	if firmware == "efi-secure" {
		firmware = "efi"
		efiSecureBootEnabled = true
	}

	confSpec.Firmware = firmware
	confSpec.BootOptions = &types.VirtualMachineBootOptions{
		EnterBIOSSetup:       types.NewBool(config.ForceBIOSSetup),
		EfiSecureBootEnabled: types.NewBool(efiSecureBootEnabled),
	}

	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachineDriver) Customize(spec types.CustomizationSpec) error {
	task, err := vm.vm.Customize(vm.driver.ctx, spec)
	if err != nil {
		return err
	}
	return task.Wait(vm.driver.ctx)
}

func (vm *VirtualMachineDriver) ResizeDisk(diskSize int64) error {
	var confSpec types.VirtualMachineConfigSpec

	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	disk, err := findDisk(devices)
	if err != nil {
		return err
	}

	disk.CapacityInKB = diskSize * 1024

	confSpec.DeviceChange = []types.BaseVirtualDeviceConfigSpec{
		&types.VirtualDeviceConfigSpec{
			Device:    disk,
			Operation: types.VirtualDeviceConfigSpecOperationEdit,
		},
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

func (vm *VirtualMachineDriver) PowerOn() error {
	task, err := vm.vm.PowerOn(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachineDriver) WaitForIP(ctx context.Context, ipNet *net.IPNet) (string, error) {
	netIP, err := vm.vm.WaitForNetIP(ctx, false)
	if err != nil {
		return "", err
	}

	for _, ips := range netIP {
		for _, ip := range ips {
			parseIP := net.ParseIP(ip)
			if ipNet != nil && !ipNet.Contains(parseIP) {
				// ip address is not in range
				continue
			}
			// default to an ipv4 addresses if no ipNet is defined
			if ipNet == nil && parseIP.To4() == nil {
				continue
			}
			return ip, nil
		}
	}

	return "", fmt.Errorf("unable to find an IP")
}

func (vm *VirtualMachineDriver) PowerOff() error {
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

func (vm *VirtualMachineDriver) IsPoweredOff() (bool, error) {
	state, err := vm.vm.PowerState(vm.driver.ctx)
	if err != nil {
		return false, err
	}

	return state == types.VirtualMachinePowerStatePoweredOff, nil
}

func (vm *VirtualMachineDriver) StartShutdown() error {
	err := vm.vm.ShutdownGuest(vm.driver.ctx)
	return err
}

func (vm *VirtualMachineDriver) WaitForShutdown(ctx context.Context, timeout time.Duration) error {
	shutdownTimer := time.After(timeout)
	for {
		off, err := vm.IsPoweredOff()
		if err != nil {
			return err
		}
		if off {
			break
		}

		select {
		case <-shutdownTimer:
			err := errors.New("Timeout while waiting for machine to shut down.")
			return err
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func (vm *VirtualMachineDriver) CreateSnapshot(name string) error {
	task, err := vm.vm.CreateSnapshot(vm.driver.ctx, name, "", false, false)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

func (vm *VirtualMachineDriver) ConvertToTemplate() error {
	return vm.vm.MarkAsTemplate(vm.driver.ctx)
}

func (vm *VirtualMachineDriver) ImportOvfToContentLibrary(ovf vcenter.OVF) error {
	err := vm.driver.restClient.Login(vm.driver.ctx)
	if err != nil {
		return err
	}

	l, err := vm.driver.FindContentLibraryByName(ovf.Target.LibraryID)
	if err != nil {
		return err
	}
	if l.library.Type != "LOCAL" {
		return fmt.Errorf("can not deploy a VM to the content library %s of type %s; "+
			"the content library must be of type LOCAL", ovf.Target.LibraryID, l.library.Type)
	}

	item, err := vm.driver.FindContentLibraryItem(l.library.ID, ovf.Spec.Name)
	if err == nil {
		// Updates existing library item
		ovf.Target.LibraryItemID = item.ID
	}

	ovf.Target.LibraryID = l.library.ID
	ovf.Source.Value = vm.vm.Reference().Value
	ovf.Source.Type = "VirtualMachine"

	vcm := vcenter.NewManager(vm.driver.restClient.client)
	_, err = vcm.CreateOVF(vm.driver.ctx, ovf)
	if err != nil {
		return err
	}

	return vm.driver.restClient.Logout(vm.driver.ctx)
}

func (vm *VirtualMachineDriver) ImportToContentLibrary(template vcenter.Template) error {
	err := vm.driver.restClient.Login(vm.driver.ctx)
	if err != nil {
		return err
	}

	l, err := vm.driver.FindContentLibraryByName(template.Library)
	if err != nil {
		return err
	}
	if l.library.Type != "LOCAL" {
		return fmt.Errorf("can not deploy a VM to the content library %s of type %s; "+
			"the content library must be of type LOCAL", template.Library, l.library.Type)
	}

	template.Library = l.library.ID
	template.SourceVM = vm.vm.Reference().Value

	if template.Placement.Cluster != "" {
		c, err := vm.driver.FindCluster(template.Placement.Cluster)
		if err != nil {
			return err
		}
		template.Placement.Cluster = c.cluster.Reference().Value
	}
	if template.Placement.Folder != "" {
		f, err := vm.driver.FindFolder(template.Placement.Folder)
		if err != nil {
			return err
		}
		template.Placement.Folder = f.folder.Reference().Value
	}
	if template.Placement.Host != "" {
		h, err := vm.driver.FindHost(template.Placement.Host)
		if err != nil {
			return err
		}
		template.Placement.Host = h.host.Reference().Value
	}
	if template.Placement.ResourcePool != "" {
		rp, err := vm.driver.FindResourcePool(template.Placement.Cluster, template.Placement.Host, template.Placement.ResourcePool)
		if err != nil {
			return err
		}
		template.Placement.ResourcePool = rp.pool.Reference().Value
	}

	if template.VMHomeStorage != nil {
		d, err := vm.driver.FindDatastore(template.VMHomeStorage.Datastore, template.Placement.Host)
		if err != nil {
			return err
		}
		template.VMHomeStorage.Datastore = d.Reference().Value
	}

	vcm := vcenter.NewManager(vm.driver.restClient.client)
	_, err = vcm.CreateTemplate(vm.driver.ctx, template)
	if err != nil {
		return err
	}

	return vm.driver.restClient.Logout(vm.driver.ctx)
}

func (vm *VirtualMachineDriver) GetDir() (string, error) {
	vmInfo, err := vm.Info("name", "layoutEx.file")
	if err != nil {
		return "", err
	}

	vmxName := fmt.Sprintf("/%s.vmx", vmInfo.Name)
	for _, file := range vmInfo.LayoutEx.File {
		if strings.Contains(file.Name, vmInfo.Name) {
			return RemoveDatastorePrefix(file.Name[:len(file.Name)-len(vmxName)]), nil
		}
	}
	return "", fmt.Errorf("cannot find '%s'", vmxName)
}

func addDisk(_ *VCenterDriver, devices object.VirtualDeviceList, config *CreateConfig) (object.VirtualDeviceList, error) {
	if len(config.Storage) == 0 {
		return nil, errors.New("no storage devices have been defined")
	}

	if len(config.DiskControllerType) == 0 {
		return nil, errors.New("no controllers have been defined")
	}

	var controllers []types.BaseVirtualController
	for _, controllerType := range config.DiskControllerType {
		var device types.BaseVirtualDevice
		var err error
		if controllerType == "nvme" {
			device, err = devices.CreateNVMEController()
		} else {
			device, err = devices.CreateSCSIController(controllerType)
		}
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
		controller, err := devices.FindDiskController(devices.Name(device))
		if err != nil {
			return nil, err
		}
		controllers = append(controllers, controller)
	}

	for _, dc := range config.Storage {
		disk := &types.VirtualDisk{
			VirtualDevice: types.VirtualDevice{
				Key: devices.NewKey(),
				Backing: &types.VirtualDiskFlatVer2BackingInfo{
					DiskMode:        string(types.VirtualDiskModePersistent),
					ThinProvisioned: types.NewBool(dc.DiskThinProvisioned),
					EagerlyScrub:    types.NewBool(dc.DiskEagerlyScrub),
				},
			},
			CapacityInKB: dc.DiskSize * 1024,
		}

		devices.AssignController(disk, controllers[dc.ControllerIndex])
		devices = append(devices, disk)
	}

	return devices, nil
}

func addNetwork(d *VCenterDriver, devices object.VirtualDeviceList, config *CreateConfig) (object.VirtualDeviceList, error) {
	if len(config.NICs) == 0 {
		return nil, errors.New("no network adapters have been defined")
	}

	for _, nic := range config.NICs {
		network, err := findNetwork(nic.Network, config.Host, d)
		if err != nil {
			return nil, err
		}

		backing, err := network.EthernetCardBackingInfo(d.ctx)
		if err != nil {
			return nil, err
		}

		device, err := object.EthernetCardTypes().CreateEthernetCard(nic.NetworkCard, backing)
		if err != nil {
			return nil, err
		}

		card := device.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
		if nic.MacAddress != "" {
			card.AddressType = string(types.VirtualEthernetCardMacTypeManual)
			card.MacAddress = nic.MacAddress
		}
		card.UptCompatibilityEnabled = nic.Passthrough

		devices = append(devices, device)
	}
	return devices, nil
}

func findNetwork(network string, host string, d *VCenterDriver) (object.NetworkReference, error) {
	if network != "" {
		var err error
		networks, err := d.FindNetworks(network)
		if err != nil {
			return nil, err
		}
		if len(networks) == 1 {
			return networks[0].network, nil
		}

		// If there are multiple networks then try to match the host
		if host != "" {
			h, err := d.FindHost(host)
			if err != nil {
				return nil, &MultipleNetworkFoundError{network, fmt.Sprintf("unable to match a network to the host %s: %s", host, err.Error())}
			}
			for _, n := range networks {
				info, err := n.Info("host")
				if err != nil {
					continue
				}
				for _, host := range info.Host {
					if h.host.Reference().Value == host.Reference().Value {
						return n.network, nil
					}
				}
			}
			return nil, &MultipleNetworkFoundError{network, fmt.Sprintf("unable to match a network to the host %s", host)}
		}

		return nil, &MultipleNetworkFoundError{network, "please provide a host to match or the network full path"}
	}

	if host != "" {
		h, err := d.FindHost(host)
		if err != nil {
			return nil, err
		}

		i, err := h.Info("network")
		if err != nil {
			return nil, err
		}

		if len(i.Network) > 1 {
			return nil, fmt.Errorf("Host has multiple networks. Specify it explicitly")
		}

		return object.NewNetwork(d.client.Client, i.Network[0]), nil
	}

	return nil, fmt.Errorf("Couldn't find network; 'host' and 'network' not specified. At least one of the two must be specified.")
}

func newVGPUProfile(vGPUProfile string) types.VirtualPCIPassthrough {
	return types.VirtualPCIPassthrough{
		VirtualDevice: types.VirtualDevice{
			DeviceInfo: &types.Description{
				Summary: "",
				Label:   fmt.Sprintf("New vGPU %v PCI device", vGPUProfile),
			},
			Backing: &types.VirtualPCIPassthroughVmiopBackingInfo{
				Vgpu: vGPUProfile,
			},
		},
	}
}

func (vm *VirtualMachineDriver) AddCdrom(controllerType string, datastoreIsoPath string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	var controller *types.VirtualController
	if controllerType == "sata" {
		c, err := vm.FindSATAController()
		if err != nil {
			return err
		}
		controller = c.GetVirtualController()
	} else {
		c, err := devices.FindIDEController("")
		if err != nil {
			return err
		}
		controller = c.GetVirtualController()
	}

	cdrom, err := vm.CreateCdrom(controller)
	if err != nil {
		return err
	}

	if datastoreIsoPath != "" {
		ds := &DatastoreIsoPath{path: datastoreIsoPath}
		if !ds.Validate() {
			return fmt.Errorf("%s is not a valid iso path", datastoreIsoPath)
		}
		if libPath, err := vm.driver.FindContentLibraryFileDatastorePath(ds.GetFilePath()); err == nil {
			datastoreIsoPath = libPath
		} else {
			log.Printf("Using %s as the datastore path", datastoreIsoPath)
		}

		devices.InsertIso(cdrom, datastoreIsoPath)
	}

	log.Printf("Creating CD-ROM on controller '%v' with iso '%v'", controller, datastoreIsoPath)
	return vm.addDevice(cdrom)
}

func (vm *VirtualMachineDriver) AddFloppy(imgPath string) error {
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

func (vm *VirtualMachineDriver) SetBootOrder(order []string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	bootOptions := types.VirtualMachineBootOptions{
		BootOrder: devices.BootOrder(order),
	}

	return vm.vm.SetBootOptions(vm.driver.ctx, &bootOptions)
}

func (vm *VirtualMachineDriver) RemoveDevice(keepFiles bool, device ...types.BaseVirtualDevice) error {
	return vm.vm.RemoveDevice(vm.driver.ctx, keepFiles, device...)
}

func (vm *VirtualMachineDriver) addDevice(device types.BaseVirtualDevice) error {
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

func (vm *VirtualMachineDriver) AddConfigParams(params map[string]string, info *types.ToolsConfigInfo) error {
	var confSpec types.VirtualMachineConfigSpec

	var ov []types.BaseOptionValue
	for k, v := range params {
		o := types.OptionValue{
			Key:   k,
			Value: v,
		}
		ov = append(ov, &o)
	}
	confSpec.ExtraConfig = ov

	confSpec.Tools = info

	if len(confSpec.ExtraConfig) > 0 || confSpec.Tools != nil {
		task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
		if err != nil {
			return err
		}

		_, err = task.WaitForResult(vm.driver.ctx, nil)
		return err
	}

	return nil
}

func (vm *VirtualMachineDriver) Export() (*nfc.Lease, error) {
	return vm.vm.Export(vm.driver.ctx)
}

func (vm *VirtualMachineDriver) CreateDescriptor(m *ovf.Manager, cdp types.OvfCreateDescriptorParams) (*types.OvfCreateDescriptorResult, error) {
	return m.CreateDescriptor(vm.driver.ctx, vm.vm, cdp)
}

func (vm *VirtualMachineDriver) NewOvfManager() *ovf.Manager {
	return ovf.NewManager(vm.vm.Client())
}

func (vm *VirtualMachineDriver) GetOvfExportOptions(m *ovf.Manager) ([]types.OvfOptionInfo, error) {
	var mgr mo.OvfManager
	err := property.DefaultCollector(vm.vm.Client()).RetrieveOne(vm.driver.ctx, m.Reference(), nil, &mgr)
	if err != nil {
		return nil, err
	}
	return mgr.OvfExportOption, nil
}

func findNetworkAdapter(l object.VirtualDeviceList) (types.BaseVirtualEthernetCard, error) {
	c := l.SelectByType((*types.VirtualEthernetCard)(nil))
	if len(c) == 0 {
		return nil, errors.New("no network adapter device found")
	}

	return c[0].(types.BaseVirtualEthernetCard), nil
}
