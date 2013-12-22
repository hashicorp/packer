package common

type DriverMock struct {
	CreateSATAControllerVM         string
	CreateSATAControllerController string
	CreateSATAControllerErr        error

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

func (d *DriverMock) IsRunning(name string) (bool, error) {
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
