package common

import "sync"

type DriverMock struct {
	sync.Mutex

	DeviceAddCdRomCalled bool
	DeviceAddCdRomName   string
	DeviceAddCdRomImage  string
	DeviceAddCdRomResult string
	DeviceAddCdRomErr    error

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

	ToolsIsoPathCalled bool
	ToolsIsoPathFlavor string
	ToolsIsoPathResult string
	ToolsIsoPathErr    error

	MacName   string
	MacReturn string
	MacError  error

	IpAddressMac    string
	IpAddressReturn string
	IpAddressError  error
}

func (d *DriverMock) DeviceAddCdRom(name string, image string) (string, error) {
	d.DeviceAddCdRomCalled = true
	d.DeviceAddCdRomName = name
	d.DeviceAddCdRomImage = image
	return d.DeviceAddCdRomResult, d.DeviceAddCdRomErr
}

func (d *DriverMock) Import(name, srcPath, dstPath string, reassignMac bool) error {
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

func (d *DriverMock) Mac(name string) (string, error) {
	d.MacName = name
	return d.MacReturn, d.MacError
}

func (d *DriverMock) IpAddress(mac string) (string, error) {
	d.IpAddressMac = mac
	return d.IpAddressReturn, d.IpAddressError
}

func (d *DriverMock) ToolsIsoPath(flavor string) (string, error) {
	d.ToolsIsoPathCalled = true
	d.ToolsIsoPathFlavor = flavor
	return d.ToolsIsoPathResult, d.ToolsIsoPathErr
}
