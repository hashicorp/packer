package common

import (
	"net"
	"sync"

	"github.com/hashicorp/packer/helper/multistep"
)

type DriverMock struct {
	sync.Mutex

	CloneCalled bool
	CloneDst    string
	CloneSrc    string
	CloneErr    error

	CompactDiskCalled bool
	CompactDiskPath   string
	CompactDiskErr    error

	CreateDiskCalled      bool
	CreateDiskOutput      string
	CreateDiskSize        string
	CreateDiskAdapterType string
	CreateDiskTypeId      string
	CreateDiskErr         error

	IsRunningCalled bool
	IsRunningPath   string
	IsRunningResult bool
	IsRunningErr    error

	CommHostCalled bool
	CommHostState  multistep.StateBag
	CommHostResult string
	CommHostErr    error

	HostAddressCalled bool
	HostAddressState  multistep.StateBag
	HostAddressResult string
	HostAddressErr    error

	HostIPCalled bool
	HostIPState  multistep.StateBag
	HostIPResult string
	HostIPErr    error

	GuestAddressCalled bool
	GuestAddressState  multistep.StateBag
	GuestAddressResult string
	GuestAddressErr    error

	GuestIPCalled bool
	GuestIPState  multistep.StateBag
	GuestIPResult string
	GuestIPErr    error

	StartCalled   bool
	StartPath     string
	StartHeadless bool
	StartErr      error

	StopCalled bool
	StopPath   string
	StopErr    error

	SuppressMessagesCalled bool
	SuppressMessagesPath   string
	SuppressMessagesErr    error

	ToolsIsoPathCalled bool
	ToolsIsoPathFlavor string
	ToolsIsoPathResult string

	ToolsInstallCalled bool
	ToolsInstallErr    error

	DhcpLeasesPathCalled bool
	DhcpLeasesPathDevice string
	DhcpLeasesPathResult string

	DhcpConfPathCalled bool
	DhcpConfPathResult string

	VmnetnatConfPathCalled bool
	VmnetnatConfPathResult string

	NetmapConfPathCalled bool
	NetmapConfPathResult string

	VerifyCalled bool
	VerifyErr    error
}

type NetworkMapperMock struct {
	NameIntoDeviceCalled int
	DeviceIntoNameCalled int
}

func (m NetworkMapperMock) NameIntoDevice(name string) (string, error) {
	m.NameIntoDeviceCalled += 1
	return "", nil
}
func (m NetworkMapperMock) DeviceIntoName(device string) (string, error) {
	m.DeviceIntoNameCalled += 1
	return "", nil
}

func (d *DriverMock) Clone(dst string, src string) error {
	d.CloneCalled = true
	d.CloneDst = dst
	d.CloneSrc = src
	return d.CloneErr
}

func (d *DriverMock) CompactDisk(path string) error {
	d.CompactDiskCalled = true
	d.CompactDiskPath = path
	return d.CompactDiskErr
}

func (d *DriverMock) CreateDisk(output string, size string, adapterType string, typeId string) error {
	d.CreateDiskCalled = true
	d.CreateDiskOutput = output
	d.CreateDiskSize = size
	d.CreateDiskAdapterType = adapterType
	d.CreateDiskTypeId = typeId
	return d.CreateDiskErr
}

func (d *DriverMock) IsRunning(path string) (bool, error) {
	d.Lock()
	defer d.Unlock()

	d.IsRunningCalled = true
	d.IsRunningPath = path
	return d.IsRunningResult, d.IsRunningErr
}

func (d *DriverMock) CommHost(state multistep.StateBag) (string, error) {
	d.CommHostCalled = true
	d.CommHostState = state
	return d.CommHostResult, d.CommHostErr
}

func MockInterface() net.Interface {
	interfaces, err := net.Interfaces()

	// Build a dummy interface due to being unable to enumerate interfaces
	if err != nil || len(interfaces) == 0 {
		return net.Interface{
			Index:        0,
			MTU:          -1,
			Name:         "dummy",
			HardwareAddr: net.HardwareAddr{0, 0, 0, 0, 0, 0},
			Flags:        net.FlagLoopback,
		}
	}

	// Find the first loopback interface
	for _, intf := range interfaces {
		if intf.Flags&net.FlagLoopback == net.FlagLoopback {
			return intf
		}
	}

	// Fall-back to just the first one
	return interfaces[0]
}

func (d *DriverMock) HostAddress(state multistep.StateBag) (string, error) {
	intf := MockInterface()
	d.HostAddressResult = intf.HardwareAddr.String()
	d.HostAddressCalled = true
	d.HostAddressState = state
	return d.HostAddressResult, d.HostAddressErr
}

func (d *DriverMock) HostIP(state multistep.StateBag) (string, error) {
	d.HostIPResult = "127.0.0.1"
	d.HostIPCalled = true
	d.HostIPState = state
	return d.HostIPResult, d.HostIPErr
}

func (d *DriverMock) GuestAddress(state multistep.StateBag) (string, error) {
	d.GuestAddressCalled = true
	d.GuestAddressState = state
	return d.GuestAddressResult, d.GuestAddressErr
}

func (d *DriverMock) GuestIP(state multistep.StateBag) (string, error) {
	d.GuestIPCalled = true
	d.GuestIPState = state
	return d.GuestIPResult, d.GuestIPErr
}

func (d *DriverMock) Start(path string, headless bool) error {
	d.StartCalled = true
	d.StartPath = path
	d.StartHeadless = headless
	return d.StartErr
}

func (d *DriverMock) Stop(path string) error {
	d.StopCalled = true
	d.StopPath = path
	return d.StopErr
}

func (d *DriverMock) SuppressMessages(path string) error {
	d.SuppressMessagesCalled = true
	d.SuppressMessagesPath = path
	return d.SuppressMessagesErr
}

func (d *DriverMock) ToolsIsoPath(flavor string) string {
	d.ToolsIsoPathCalled = true
	d.ToolsIsoPathFlavor = flavor
	return d.ToolsIsoPathResult
}

func (d *DriverMock) ToolsInstall() error {
	d.ToolsInstallCalled = true
	return d.ToolsInstallErr
}

func (d *DriverMock) DhcpLeasesPath(device string) string {
	d.DhcpLeasesPathCalled = true
	d.DhcpLeasesPathDevice = device
	return d.DhcpLeasesPathResult
}

func (d *DriverMock) DhcpConfPath(device string) string {
	d.DhcpConfPathCalled = true
	return d.DhcpConfPathResult
}

func (d *DriverMock) VmnetnatConfPath(device string) string {
	d.VmnetnatConfPathCalled = true
	return d.VmnetnatConfPathResult
}

func (d *DriverMock) NetmapConfPath() string {
	d.NetmapConfPathCalled = true
	return d.NetmapConfPathResult
}

func (d *DriverMock) Verify() error {
	d.VerifyCalled = true
	return d.VerifyErr
}

func (d *DriverMock) GetVmwareDriver() VmwareDriver {
	var state VmwareDriver
	state.DhcpLeasesPath = func(string) string {
		return "/path/to/dhcp.leases"
	}
	state.DhcpConfPath = func(string) string {
		return "/path/to/dhcp.conf"
	}
	state.VmnetnatConfPath = func(string) string {
		return "/path/to/vmnetnat.conf"
	}
	state.NetworkMapper = func() (NetworkNameMapper, error) {
		return NetworkMapperMock{}, nil
	}
	return state
}
