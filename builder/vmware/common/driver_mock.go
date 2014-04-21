package common

import (
	"sync"

	"github.com/mitchellh/multistep"
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

	CreateDiskCalled bool
	CreateDiskOutput string
	CreateDiskSize   string
	CreateDiskTypeId string
	CreateDiskErr    error

	IsRunningCalled bool
	IsRunningVmId   string
	IsRunningResult bool
	IsRunningErr    error

	SSHAddressCalled bool
	SSHAddressState  multistep.StateBag
	SSHAddressResult string
	SSHAddressErr    error

	StartCalled   bool
	StartVmId     string
	StartHeadless bool
	StartErr      error

	StopCalled bool
	StopVmId   string
	StopErr    error

	SuppressMessagesCalled bool
	SuppressMessagesPath   string
	SuppressMessagesErr    error

	ToolsIsoPathCalled bool
	ToolsIsoPathFlavor string
	ToolsIsoPathResult string

	DhcpLeasesPathCalled bool
	DhcpLeasesPathDevice string
	DhcpLeasesPathResult string

	VerifyCalled bool
	VerifyErr    error
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

func (d *DriverMock) CreateDisk(output string, size string, typeId string) error {
	d.CreateDiskCalled = true
	d.CreateDiskOutput = output
	d.CreateDiskSize = size
	d.CreateDiskTypeId = typeId
	return d.CreateDiskErr
}

func (d *DriverMock) IsRunning(vmId string) (bool, error) {
	d.Lock()
	defer d.Unlock()

	d.IsRunningCalled = true
	d.IsRunningVmId = vmId
	return d.IsRunningResult, d.IsRunningErr
}

func (d *DriverMock) SSHAddress(state multistep.StateBag) (string, error) {
	d.SSHAddressCalled = true
	d.SSHAddressState = state
	return d.SSHAddressResult, d.SSHAddressErr
}

func (d *DriverMock) Start(vmId string, headless bool) error {
	d.StartCalled = true
	d.StartVmId = vmId
	d.StartHeadless = headless
	return d.StartErr
}

func (d *DriverMock) Stop(vmId string) error {
	d.StopCalled = true
	d.StopVmId = vmId
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

func (d *DriverMock) DhcpLeasesPath(device string) string {
	d.DhcpLeasesPathCalled = true
	d.DhcpLeasesPathDevice = device
	return d.DhcpLeasesPathResult
}

func (d *DriverMock) Verify() error {
	d.VerifyCalled = true
	return d.VerifyErr
}
