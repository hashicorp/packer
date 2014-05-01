package common

import "sync"

type DriverMock struct {
	sync.Mutex

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

	MacName   string
	MacReturn string
	MacError  error

	IpAddressMac    string
	IpAddressReturn string
	IpAddressError  error
}

func (d *DriverMock) Import(name, srcPath, dstPath string) error {
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

func (d *DriverMock) Mac(name string) (string, error) {
	d.MacName = name
	return d.MacReturn, d.MacError
}

func (d *DriverMock) IpAddress(mac string) (string, error) {
	d.IpAddressMac = mac
	return d.IpAddressReturn, d.IpAddressError
}
