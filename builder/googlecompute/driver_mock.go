package googlecompute

// DriverMock is a Driver implementation that is a mocked out so that
// it can be used for tests.
type DriverMock struct {
	CreateImageName  string
	CreateImageDesc  string
	CreateImageURL   string
	CreateImageErrCh <-chan error

	DeleteImageName  string
	DeleteImageErrCh <-chan error

	DeleteInstanceZone  string
	DeleteInstanceName  string
	DeleteInstanceErrCh <-chan error
	DeleteInstanceErr   error

	GetNatIPZone   string
	GetNatIPName   string
	GetNatIPResult string
	GetNatIPErr    error

	RunInstanceConfig *InstanceConfig
	RunInstanceErrCh  <-chan error
	RunInstanceErr    error

	WaitForInstanceState string
	WaitForInstanceZone  string
	WaitForInstanceName  string
	WaitForInstanceErrCh <-chan error
}

func (d *DriverMock) CreateImage(name, description, url string) <-chan error {
	d.CreateImageName = name
	d.CreateImageDesc = description
	d.CreateImageURL = url

	resultCh := d.CreateImageErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh
}

func (d *DriverMock) DeleteImage(name string) <-chan error {
	d.DeleteImageName = name

	resultCh := d.DeleteImageErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh
}

func (d *DriverMock) DeleteInstance(zone, name string) (<-chan error, error) {
	d.DeleteInstanceZone = zone
	d.DeleteInstanceName = name

	resultCh := d.DeleteInstanceErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh, d.DeleteInstanceErr
}

func (d *DriverMock) GetNatIP(zone, name string) (string, error) {
	d.GetNatIPZone = zone
	d.GetNatIPName = name
	return d.GetNatIPResult, d.GetNatIPErr
}

func (d *DriverMock) RunInstance(c *InstanceConfig) (<-chan error, error) {
	d.RunInstanceConfig = c

	resultCh := d.RunInstanceErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh, d.RunInstanceErr
}

func (d *DriverMock) WaitForInstance(state, zone, name string) <-chan error {
	d.WaitForInstanceState = state
	d.WaitForInstanceZone = zone
	d.WaitForInstanceName = name

	resultCh := d.WaitForInstanceErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh
}
