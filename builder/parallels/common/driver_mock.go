package common

import "sync"

type DriverMock struct {
	sync.Mutex

	CompactDiskCalled bool
	CompactDiskPath   string
	CompactDiskErr    error

	DeviceAddCDROMCalled bool
	DeviceAddCDROMName   string
	DeviceAddCDROMImage  string
	DeviceAddCDROMResult string
	DeviceAddCDROMErr    error

	DiskPathCalled bool
	DiskPathName   string
	DiskPathResult string
	DiskPathErr    error

	ImportCalled  bool
	ImportName    string
	ImportSrcPath string
	ImportDstPath string
	ImportErr     error

	IsRunningName   string
	IsRunningReturn bool
	IsRunningErr    error

	StopName string
	StopErr  error

	PrlctlCalls [][]string
	PrlctlErrs  []error

	VerifyCalled bool
	VerifyErr    error

	VersionCalled bool
	VersionResult string
	VersionErr    error

	SendKeyScanCodesCalls [][]string
	SendKeyScanCodesErrs  []error

	SetDefaultConfigurationCalled bool
	SetDefaultConfigurationError  error

	ToolsISOPathCalled bool
	ToolsISOPathFlavor string
	ToolsISOPathResult string
	ToolsISOPathErr    error

	MACName   string
	MACReturn string
	MACError  error

	IPAddressMAC    string
	IPAddressReturn string
	IPAddressError  error
}

func (d *DriverMock) CompactDisk(path string) error {
	d.CompactDiskCalled = true
	d.CompactDiskPath = path
	return d.CompactDiskErr
}

func (d *DriverMock) DeviceAddCDROM(name string, image string) (string, error) {
	d.DeviceAddCDROMCalled = true
	d.DeviceAddCDROMName = name
	d.DeviceAddCDROMImage = image
	return d.DeviceAddCDROMResult, d.DeviceAddCDROMErr
}

func (d *DriverMock) DiskPath(name string) (string, error) {
	d.DiskPathCalled = true
	d.DiskPathName = name
	return d.DiskPathResult, d.DiskPathErr
}

func (d *DriverMock) Import(name, srcPath, dstPath string, reassignMAC bool) error {
	d.ImportCalled = true
	d.ImportName = name
	d.ImportSrcPath = srcPath
	d.ImportDstPath = dstPath
	return d.ImportErr
}

func (d *DriverMock) IsRunning(name string) (bool, error) {
	d.Lock()
	defer d.Unlock()

	d.IsRunningName = name
	return d.IsRunningReturn, d.IsRunningErr
}

func (d *DriverMock) Stop(name string) error {
	d.StopName = name
	return d.StopErr
}

func (d *DriverMock) Prlctl(args ...string) error {
	d.PrlctlCalls = append(d.PrlctlCalls, args)

	if len(d.PrlctlErrs) >= len(d.PrlctlCalls) {
		return d.PrlctlErrs[len(d.PrlctlCalls)-1]
	}
	return nil
}

func (d *DriverMock) Verify() error {
	d.VerifyCalled = true
	return d.VerifyErr
}

func (d *DriverMock) Version() (string, error) {
	d.VersionCalled = true
	return d.VersionResult, d.VersionErr
}

func (d *DriverMock) SendKeyScanCodes(name string, scancodes ...string) error {
	d.SendKeyScanCodesCalls = append(d.SendKeyScanCodesCalls, scancodes)

	if len(d.SendKeyScanCodesErrs) >= len(d.SendKeyScanCodesCalls) {
		return d.SendKeyScanCodesErrs[len(d.SendKeyScanCodesCalls)-1]
	}
	return nil
}

func (d *DriverMock) SetDefaultConfiguration(name string) error {
	d.SetDefaultConfigurationCalled = true
	return d.SetDefaultConfigurationError
}

func (d *DriverMock) MAC(name string) (string, error) {
	d.MACName = name
	return d.MACReturn, d.MACError
}

func (d *DriverMock) IPAddress(mac string) (string, error) {
	d.IPAddressMAC = mac
	return d.IPAddressReturn, d.IPAddressError
}

func (d *DriverMock) ToolsISOPath(flavor string) (string, error) {
	d.ToolsISOPathCalled = true
	d.ToolsISOPathFlavor = flavor
	return d.ToolsISOPathResult, d.ToolsISOPathErr
}
