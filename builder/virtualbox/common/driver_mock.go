package common

import "sync"

type DriverMock struct {
	sync.Mutex

	CreateSATAControllerVM         string
	CreateSATAControllerController string
	CreateSATAControllerErr        error

	DeleteCalled bool
	DeleteName   string
	DeleteErr    error

	ImportCalled bool
	ImportName   string
	ImportPath   string
	ImportOpts   string
	ImportErr    error

	IsRunningName   string
	IsRunningReturn bool
	IsRunningErr    error

	ResetCalled   bool
	ResetPath     string
	ResetHeadless bool
	ResetErr      error

	StopName string
	StopErr  error

	SuppressMessagesCalled bool
	SuppressMessagesErr    error

	VBoxManageCalls [][]string
	VBoxManageErrs  []error

	VerifyCalled bool
	VerifyErr    error

	VersionCalled bool
	VersionResult string
	VersionErr    error
}

func (d *DriverMock) CreateSATAController(vm string, controller string) error {
	d.CreateSATAControllerVM = vm
	d.CreateSATAControllerController = vm
	return d.CreateSATAControllerErr
}

func (d *DriverMock) Delete(name string) error {
	d.DeleteCalled = true
	d.DeleteName = name
	return d.DeleteErr
}

func (d *DriverMock) Import(name, path, opts string) error {
	d.ImportCalled = true
	d.ImportName = name
	d.ImportPath = path
	d.ImportOpts = opts
	return d.ImportErr
}

func (d *DriverMock) IsRunning(name string) (bool, error) {
	d.Lock()
	defer d.Unlock()

	d.IsRunningName = name
	return d.IsRunningReturn, d.IsRunningErr
}

func (d *DriverMock) Reset(path string, headless bool) error {
	d.ResetCalled = true
	d.ResetPath = path
	d.ResetHeadless = headless
	return d.ResetErr
}

func (d *DriverMock) Stop(name string) error {
	d.StopName = name
	return d.StopErr
}

func (d *DriverMock) SuppressMessages() error {
	d.SuppressMessagesCalled = true
	return d.SuppressMessagesErr
}

func (d *DriverMock) VBoxManage(args ...string) error {
	d.VBoxManageCalls = append(d.VBoxManageCalls, args)

	if len(d.VBoxManageErrs) >= len(d.VBoxManageCalls) {
		return d.VBoxManageErrs[len(d.VBoxManageCalls)-1]
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
