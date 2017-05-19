package common

import (
	"fmt"
	"sync"
)

type DriverMock struct {
	sync.Mutex

	CloneVirtualMachineCalled         bool
	CloneVirtualMachineSrcVmName      string
	CloneVirtualMachineDstVmName      string
	CloneVirtualMachineSrcFolder      string
	CloneVirtualMachineDstFolder      string
	CloneVirtualMachineSrcDatacenter  string
	CloneVirtualMachineDstDatastore   string
	CloneVirtualMachineCpu            uint
	CloneVirtualMachineRam            uint
	CloneVirtualMachineNetworkName    string
	CloneVirtualMachineNetworkAdapter string
	CloneVirtualMachineAnnotation     string
	CloneVirtualMachineErr            error

	CreateVirtualMachineCalled         bool
	CreateVirtualMachineVmName         string
	CreateVirtualMachineFolder         string
	CreateVirtualMachineDatastore      string
	CreateVirtualMachineCpu            uint
	CreateVirtualMachineRam            uint
	CreateVirtualMachineDiskSize       uint
	CreateVirtualMachineDiskThick      bool
	CreateVirtualMachineGuestType      string
	CreateVirtualMachineNetworkName    string
	CreateVirtualMachineNetworkAdapter string
	CreateVirtualMachineAnnotation     string
	CreateVirtualMachineErr            error

	CreateDiskCalled bool
	CreateDiskSize   uint
	CreateDiskThick  bool
	CreateDiskOutput []uint
	CreateDiskErr    error

	IsRunningCalled bool
	IsRunningPath   string
	IsRunningResult bool
	IsRunningErr    error

	StartCalled bool
	StartErr    error

	StopCalled bool
	StopErr    error

	DestroyCalled bool
	DestroyErr    error

	ToolsInstallCalled bool
	ToolsInstallErr    error

	UploadCalled         bool
	UploadLocalPath      string
	UploadRemoteFilename string
	UploadResult         string
	UploadErr            error

	ExportVirtualMachineCalled    bool
	ExportVirtualMachineLocalPath string
	ExportVirtualMachineFormat    string
	ExportVirtualMachineOptions   []string
	ExportVirtualMachineErr       error

	IsDestroyedCalled bool
	IsDestroyedResult bool
	IsDestroyedErr    error

	IsStoppedCalled bool
	IsStoppedResult bool
	IsStoppedErr    error

	AddFloppyCalled   bool
	AddFloppyFilename string
	AddFloppyOutput   string
	AddFloppyErr      error

	RemoveFloppyCalled bool
	RemoveFloppyDevice string
	RemoveFloppyErr    error

	MountISOCalled   bool
	MountISOFilename string
	MountISOOutput   string
	MountISOErr      error

	UnmountISOCalled bool
	UnmountISODevice string
	UnmountISOErr    error

	VMChangeCalled bool
	VMChangeOption []string
	VMChangeErr    error

	VNCDisableCalled bool
	VNCDisableErr    error

	VNCEnableCalled       bool
	VNCEnableVncPassword  string
	VNCEnableVncPortMin   uint
	VNCEnableVncPortMax   uint
	VNCEnableOutputString string
	VNCEnableOutputUint   uint
	VNCEnableErr          error

	GuestIPCalled bool
	GuestIPOutput string
	GuestIPErr    error

	VerifyCalled bool
	VerifyErr    error
}

func (d *DriverMock) CloneVirtualMachine(srcVmName string, dstVmName string, srcFolder string, dstFolder string, srcDatacenter string, dstDatastore string, cpu uint, ram uint, networkName string, networkAdapter string, annotation string) error {
	d.CloneVirtualMachineCalled = true
	d.CloneVirtualMachineSrcVmName = srcVmName
	d.CloneVirtualMachineDstVmName = dstVmName
	d.CloneVirtualMachineSrcFolder = srcFolder
	d.CloneVirtualMachineDstFolder = dstFolder
	d.CloneVirtualMachineSrcDatacenter = srcDatacenter
	d.CloneVirtualMachineDstDatastore = dstDatastore
	d.CloneVirtualMachineCpu = cpu
	d.CloneVirtualMachineRam = ram
	d.CloneVirtualMachineNetworkName = networkName
	d.CloneVirtualMachineNetworkAdapter = networkAdapter
	d.CloneVirtualMachineAnnotation = annotation
	return d.CloneVirtualMachineErr
}

func (d *DriverMock) CreateVirtualMachine(vmName string, folder string, datastore string, cpu uint, ram uint, diskSize uint, diskThick bool, guestType string, networkName string, networkAdapter string, annotation string) error {
	d.CreateVirtualMachineCalled = true
	d.CreateVirtualMachineVmName = vmName
	d.CreateVirtualMachineFolder = folder
	d.CreateVirtualMachineDatastore = datastore
	d.CreateVirtualMachineCpu = cpu
	d.CreateVirtualMachineRam = ram
	d.CreateVirtualMachineDiskSize = diskSize
	d.CreateVirtualMachineDiskThick = diskThick
	d.CreateVirtualMachineGuestType = guestType
	d.CreateVirtualMachineNetworkName = networkName
	d.CreateVirtualMachineNetworkAdapter = networkAdapter
	d.CreateVirtualMachineAnnotation = annotation
	return d.CreateVirtualMachineErr
}

func (d *DriverMock) CreateDisk(diskSize uint, diskThick bool) error {
	d.CreateDiskCalled = true
	d.CreateDiskSize = diskSize
	d.CreateDiskThick = diskThick
	d.CreateDiskOutput = append(d.CreateDiskOutput, diskSize)
	return d.CreateDiskErr
}

func (d *DriverMock) IsRunning() (bool, error) {
	d.Lock()
	defer d.Unlock()

	d.IsRunningCalled = true
	//TODO: is the correct value ?
	d.IsRunningResult = d.StartCalled
	return d.IsRunningResult, d.IsRunningErr
}

func (d *DriverMock) Start() error {
	d.StartCalled = true
	return d.StartErr
}

func (d *DriverMock) Stop() error {
	d.StopCalled = true
	return d.StopErr
}

func (d *DriverMock) ToolsInstall() error {
	d.ToolsInstallCalled = true
	return d.ToolsInstallErr
}

func (d *DriverMock) Upload(localPath string, remoteFilename string) (string, error) {
	d.UploadCalled = true
	d.UploadLocalPath = localPath
	d.UploadRemoteFilename = remoteFilename
	d.UploadResult = fmt.Sprintf("/datastore/%s", remoteFilename)
	return d.UploadResult, d.UploadErr
}

func (d *DriverMock) ExportVirtualMachine(localpath string, format string, options []string) error {
	d.ExportVirtualMachineCalled = true
	d.ExportVirtualMachineLocalPath = localpath
	d.ExportVirtualMachineFormat = format
	d.ExportVirtualMachineOptions = options
	return d.ExportVirtualMachineErr
}

func (d *DriverMock) Destroy() error {
	d.DestroyCalled = true
	return d.DestroyErr
}

func (d *DriverMock) IsDestroyed() (bool, error) {
	d.IsDestroyedCalled = true
	d.IsDestroyedResult = d.DestroyCalled
	return d.IsDestroyedResult, d.IsDestroyedErr
}

func (d *DriverMock) IsStopped() (bool, error) {
	d.IsStoppedCalled = true
	d.IsStoppedResult = true
	return d.IsStoppedResult, d.IsStoppedErr
}

func (d *DriverMock) AddFloppy(floppyFilename string) (string, error) {
	d.AddFloppyCalled = true
	d.AddFloppyFilename = floppyFilename
	d.AddFloppyOutput = "floppy1"
	return d.AddFloppyOutput, d.AddFloppyErr
}

func (d *DriverMock) RemoveFloppy(floppyDevice string) error {
	d.RemoveFloppyCalled = true
	d.RemoveFloppyDevice = floppyDevice
	return d.RemoveFloppyErr
}

func (d *DriverMock) MountISO(isoFilename string) (string, error) {
	d.MountISOCalled = true
	d.MountISOFilename = isoFilename
	d.MountISOOutput = "cdrom1"
	return d.MountISOOutput, d.MountISOErr
}

func (d *DriverMock) UnmountISO(cdromDevice string) error {
	d.UnmountISOCalled = true
	d.UnmountISODevice = cdromDevice
	return d.UnmountISOErr
}

func (d *DriverMock) VMChange(change string) error {
	d.VMChangeCalled = true
	d.VMChangeOption = append(d.VMChangeOption, change)
	return d.VMChangeErr
}

func (d *DriverMock) VNCDisable() error {
	d.VNCDisableCalled = true
	return d.VNCDisableErr
}

func (d *DriverMock) VNCEnable(vncpassword string, vncportmin uint, vncportmax uint) (string, uint, error) {
	d.VNCEnableCalled = true
	d.VNCEnableVncPassword = vncpassword
	d.VNCEnableVncPortMin = vncportmin
	d.VNCEnableVncPortMax = vncportmax
	d.VNCEnableOutputUint = d.VNCEnableVncPortMin - 1
	return d.VNCEnableOutputString, d.VNCEnableOutputUint, d.VNCEnableErr
}

func (d *DriverMock) GuestIP() (string, error) {
	d.GuestIPCalled = true
	return d.GuestIPOutput, d.GuestIPErr
}

func (d *DriverMock) Verify() error {
	d.VerifyCalled = true
	return d.VerifyErr
}
