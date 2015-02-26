package iso

import (
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
)

type RemoteDriverMock struct {
	vmwcommon.DriverMock

	UploadISOCalled bool
	UploadISOPath   string
	UploadISOResult string
	UploadISOErr    error

	RegisterCalled bool
	RegisterPath   string
	RegisterErr    error

	DestroyCalled bool
	DestroyErr    error

	IsDestroiedCalled bool
	IsDestroiedResult bool
	IsDestroiedErr    error

	uploadErr error

	ReloadVMErr error
}

func (d *RemoteDriverMock) UploadISO(path string, checksum string, checksumType string) (string, error) {
	d.UploadISOCalled = true
	d.UploadISOPath = path
	return d.UploadISOResult, d.UploadISOErr
}

func (d *RemoteDriverMock) Register(path string) error {
	d.RegisterCalled = true
	d.RegisterPath = path
	return d.RegisterErr
}

func (d *RemoteDriverMock) Destroy() error {
	d.DestroyCalled = true
	return d.DestroyErr
}

func (d *RemoteDriverMock) IsDestroied() (bool, error) {
	d.DestroyCalled = true
	return d.IsDestroiedResult, d.IsDestroiedErr
}

func (d *RemoteDriverMock) upload(dst, src string) error {
	return d.uploadErr
}

func (d *RemoteDriverMock) ReloadVM() error {
	return d.ReloadVMErr
}
