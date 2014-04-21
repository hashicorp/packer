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

	UnregisterCalled bool
	UnregisterVmId   string
	UnregisterErr    error
}

func (d *RemoteDriverMock) UploadISO(path string) (string, error) {
	d.UploadISOCalled = true
	d.UploadISOPath = path
	return d.UploadISOResult, d.UploadISOErr
}

func (d *RemoteDriverMock) Register(path string) (string, error) {
	d.RegisterCalled = true
	d.RegisterPath = path
	return "1", d.RegisterErr
}

func (d *RemoteDriverMock) Unregister(vmId string) error {
	d.UnregisterCalled = true
	d.UnregisterVmId = vmId
	return d.UnregisterErr
}
