package common

import "sync"

type DriverMock struct {
	sync.Mutex

	CreateSATAControllerVM         string
	CreateSATAControllerController string
	CreateSATAControllerErr        error

	CreateSCSIControllerVM         string
	CreateSCSIControllerController string
	CreateSCSIControllerErr        error

	DeleteCalled bool
	DeleteName   string
	DeleteErr    error

	ImportCalled bool
	ImportName   string
	ImportPath   string
	ImportFlags  []string
	ImportErr    error

	IsoCalled bool
	IsoErr    error

	IsRunningName   string
	IsRunningReturn bool
	IsRunningErr    error

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

func (d *DriverMock) CreateSCSIController(vm string, controller string) error {
	d.CreateSCSIControllerVM = vm
	d.CreateSCSIControllerController = vm
	return d.CreateSCSIControllerErr
}

func (d *DriverMock) Delete(name string) error {
	d.DeleteCalled = true
	d.DeleteName = name
	return d.DeleteErr
}

func (d *DriverMock) Import(name string, path string, flags []string) error {
	d.ImportCalled = true
	d.ImportName = name
	d.ImportPath = path
	d.ImportFlags = flags
	return d.ImportErr
}

func (d *DriverMock) Iso() (string, error) {
	d.IsoCalled = true
	return "", d.IsoErr
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
