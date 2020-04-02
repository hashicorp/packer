package common

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

func (d *RemoteDriverMock) upload(dst, src string) error {
	return d.UploadErr
}

func (d *RemoteDriverMock) Download(src, dst string) error {
	return d.DownloadErr
}

func (d *RemoteDriverMock) RemoveCache(localPath string) error {
	return nil
}

func (d *RemoteDriverMock) ReloadVM() error {
	return d.ReloadVMErr
}
