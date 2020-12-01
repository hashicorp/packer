package common

import (
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type RemoteDriverMock struct {
	DriverMock

	UploadISOCalled bool
	UploadISOPath   string
	UploadISOResult string
	UploadISOErr    error

	RegisterCalled bool
	RegisterPath   string
	RegisterErr    error

	UnregisterCalled bool
	UnregisterPath   string
	UnregisterErr    error

	DestroyCalled bool
	DestroyErr    error

	IsDestroyedCalled bool
	IsDestroyedResult bool
	IsDestroyedErr    error

	UploadErr   error
	DownloadErr error

	RemovedCachePath string
	CacheRemoved     bool

	ReturnValDirExists bool

	ReloadVMErr error

	outputDir string
}

func (d *RemoteDriverMock) UploadISO(path string, checksum string, ui packersdk.Ui) (string, error) {
	d.UploadISOCalled = true
	d.UploadISOPath = path
	return d.UploadISOResult, d.UploadISOErr
}

func (d *RemoteDriverMock) Register(path string) error {
	d.RegisterCalled = true
	d.RegisterPath = path
	return d.RegisterErr
}

func (d *RemoteDriverMock) Unregister(path string) error {
	d.UnregisterCalled = true
	d.UnregisterPath = path
	return d.UnregisterErr
}

func (d *RemoteDriverMock) Destroy() error {
	d.DestroyCalled = true
	return d.DestroyErr
}

func (d *RemoteDriverMock) IsDestroyed() (bool, error) {
	d.DestroyCalled = true
	return d.IsDestroyedResult, d.IsDestroyedErr
}

func (d *RemoteDriverMock) upload(dst, src string, ui packersdk.Ui) error {
	return d.UploadErr
}

func (d *RemoteDriverMock) Download(src, dst string) error {
	return d.DownloadErr
}

func (d *RemoteDriverMock) RemoveCache(localPath string) error {
	d.RemovedCachePath = localPath
	d.CacheRemoved = true
	return nil
}

func (d *RemoteDriverMock) ReloadVM() error {
	return d.ReloadVMErr
}

// the following functions satisfy the Outputdir interface

func (d *RemoteDriverMock) DirExists() (bool, error) {
	return d.ReturnValDirExists, nil
}

func (d *RemoteDriverMock) ListFiles() ([]string, error) {
	return []string{}, nil
}

func (d *RemoteDriverMock) MkdirAll() error {
	return nil
}

func (d *RemoteDriverMock) Remove(string) error {
	return nil
}

func (d *RemoteDriverMock) RemoveAll() error {
	return nil
}

func (d *RemoteDriverMock) SetOutputDir(s string) {
	d.outputDir = s
}

func (d *RemoteDriverMock) String() string {
	return d.outputDir
}
