package driver

import (
	"context"
	"net"
	"time"

	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VirtualMachineMock struct {
	DestroyError  error
	DestroyCalled bool

	ConfigureError          error
	ConfigureCalled         bool
	ConfigureHardwareConfig *HardwareConfig

	FindSATAControllerCalled bool
	FindSATAControllerErr    error

	AddSATAControllerCalled bool
	AddSATAControllerErr    error

	AddCdromCalled      bool
	AddCdromCalledTimes int
	AddCdromErr         error
	AddCdromTypes       []string
	AddCdromPaths       []string
}

func (vm *VirtualMachineMock) Info(params ...string) (*mo.VirtualMachine, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) Devices() (object.VirtualDeviceList, error) {
	return object.VirtualDeviceList{}, nil
}

func (vm *VirtualMachineMock) Clone(ctx context.Context, config *CloneConfig) (VirtualMachine, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) updateVAppConfig(ctx context.Context, newProps map[string]string) (*types.VmConfigSpec, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) AddPublicKeys(ctx context.Context, publicKeys string) error {
	return nil
}

func (vm *VirtualMachineMock) Properties(ctx context.Context) (*mo.VirtualMachine, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) Destroy() error {
	vm.DestroyCalled = true
	if vm.DestroyError != nil {
		return vm.DestroyError
	}
	return nil
}

func (vm *VirtualMachineMock) Configure(config *HardwareConfig) error {
	vm.ConfigureCalled = true
	vm.ConfigureHardwareConfig = config
	if vm.ConfigureError != nil {
		return vm.ConfigureError
	}
	return nil
}

func (vm *VirtualMachineMock) Customize(spec types.CustomizationSpec) error {
	return nil
}

func (vm *VirtualMachineMock) ResizeDisk(diskSize int64) error {
	return nil
}

func (vm *VirtualMachineMock) PowerOn() error {
	return nil
}

func (vm *VirtualMachineMock) WaitForIP(ctx context.Context, ipNet *net.IPNet) (string, error) {
	return "", nil
}

func (vm *VirtualMachineMock) PowerOff() error {
	return nil
}

func (vm *VirtualMachineMock) IsPoweredOff() (bool, error) {
	return false, nil
}

func (vm *VirtualMachineMock) StartShutdown() error {
	return nil
}

func (vm *VirtualMachineMock) WaitForShutdown(ctx context.Context, timeout time.Duration) error {
	return nil
}

func (vm *VirtualMachineMock) CreateSnapshot(name string) error {
	return nil
}

func (vm *VirtualMachineMock) ConvertToTemplate() error {
	return nil
}

func (vm *VirtualMachineMock) ImportOvfToContentLibrary(ovf vcenter.OVF) error {
	return nil
}

func (vm *VirtualMachineMock) ImportToContentLibrary(template vcenter.Template) error {
	return nil
}

func (vm *VirtualMachineMock) GetDir() (string, error) {
	return "", nil
}

func (vm *VirtualMachineMock) AddCdrom(cdromType string, isoPath string) error {
	vm.AddCdromCalledTimes++
	vm.AddCdromCalled = true
	vm.AddCdromTypes = append(vm.AddCdromTypes, cdromType)
	vm.AddCdromPaths = append(vm.AddCdromPaths, isoPath)
	return vm.AddCdromErr
}

func (vm *VirtualMachineMock) AddFloppy(imgPath string) error {
	return nil
}

func (vm *VirtualMachineMock) SetBootOrder(order []string) error {
	return nil
}

func (vm *VirtualMachineMock) RemoveDevice(keepFiles bool, device ...types.BaseVirtualDevice) error {
	return nil
}

func (vm *VirtualMachineMock) addDevice(device types.BaseVirtualDevice) error {
	return nil
}

func (vm *VirtualMachineMock) AddConfigParams(params map[string]string, info *types.ToolsConfigInfo) error {
	return nil
}

func (vm *VirtualMachineMock) Export() (*nfc.Lease, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) CreateDescriptor(m *ovf.Manager, cdp types.OvfCreateDescriptorParams) (*types.OvfCreateDescriptorResult, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) NewOvfManager() *ovf.Manager {
	return nil
}

func (vm *VirtualMachineMock) GetOvfExportOptions(m *ovf.Manager) ([]types.OvfOptionInfo, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) AddSATAController() error {
	vm.AddSATAControllerCalled = true
	return vm.AddSATAControllerErr
}

func (vm *VirtualMachineMock) FindSATAController() (*types.VirtualAHCIController, error) {
	vm.FindSATAControllerCalled = true
	return nil, vm.FindSATAControllerErr
}

func (vm *VirtualMachineMock) CreateCdrom(c *types.VirtualController) (*types.VirtualCdrom, error) {
	return nil, nil
}

func (vm *VirtualMachineMock) RemoveCdroms() error {
	return nil
}

func (vm *VirtualMachineMock) EjectCdroms() error {
	return nil
}
