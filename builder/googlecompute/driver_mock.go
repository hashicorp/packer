package googlecompute

// DriverMock is a Driver implementation that is a mocked out so that
// it can be used for tests.
type DriverMock struct {
	ImageExistsName   string
	ImageExistsResult bool

	CreateImageName      string
	CreateImageDesc      string
	CreateImageFamily    string
	CreateImageZone      string
	CreateImageDisk      string
	CreateImageProjectId string
	CreateImageSizeGb    int64
	CreateImageErrCh     <-chan error
	CreateImageResultCh  <-chan Image

	DeleteImageName  string
	DeleteImageErrCh <-chan error

	DeleteInstanceZone  string
	DeleteInstanceName  string
	DeleteInstanceErrCh <-chan error
	DeleteInstanceErr   error

	DeleteDiskZone  string
	DeleteDiskName  string
	DeleteDiskErrCh <-chan error
	DeleteDiskErr   error

	GetNatIPZone   string
	GetNatIPName   string
	GetNatIPResult string
	GetNatIPErr    error

	GetInternalIPZone   string
	GetInternalIPName   string
	GetInternalIPResult string
	GetInternalIPErr    error
	
	GetSerialPortOutputZone   string
	GetSerialPortOutputName   string
	GetSerialPortOutputResult string
	GetSerialPortOutputErr    error

	RunInstanceConfig *InstanceConfig
	RunInstanceErrCh  <-chan error
	RunInstanceErr    error

	WaitForInstanceState string
	WaitForInstanceZone  string
	WaitForInstanceName  string
	WaitForInstanceErrCh <-chan error
}

func (d *DriverMock) ImageExists(name string) bool {
	d.ImageExistsName = name
	return d.ImageExistsResult
}

func (d *DriverMock) CreateImage(name, description, family, zone, disk string) (<-chan Image, <-chan error) {
	d.CreateImageName = name
	d.CreateImageDesc = description
	d.CreateImageFamily = family
	d.CreateImageZone = zone
	d.CreateImageDisk = disk
	if d.CreateImageSizeGb == 0 {
		d.CreateImageSizeGb = 10
	}
	if d.CreateImageProjectId == "" {
		d.CreateImageProjectId = "test"
	}

	resultCh := d.CreateImageResultCh
	if resultCh == nil {
		ch := make(chan Image, 1)
		ch <- Image{
			Name:      name,
			ProjectId: d.CreateImageProjectId,
			SizeGb:    d.CreateImageSizeGb,
		}
		close(ch)
		resultCh = ch
	}

	errCh := d.CreateImageErrCh
	if errCh == nil {
		ch := make(chan error)
		close(ch)
		errCh = ch
	}

	return resultCh, errCh
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

func (d *DriverMock) DeleteDisk(zone, name string) (<-chan error, error) {
	d.DeleteDiskZone = zone
	d.DeleteDiskName = name

	resultCh := d.DeleteDiskErrCh
	if resultCh == nil {
		ch := make(chan error)
		close(ch)
		resultCh = ch
	}

	return resultCh, d.DeleteDiskErr
}

func (d *DriverMock) GetNatIP(zone, name string) (string, error) {
	d.GetNatIPZone = zone
	d.GetNatIPName = name
	return d.GetNatIPResult, d.GetNatIPErr
}

func (d *DriverMock) GetInternalIP(zone, name string) (string, error) {
	d.GetInternalIPZone = zone
	d.GetInternalIPName = name
	return d.GetInternalIPResult, d.GetInternalIPErr
}

func (d *DriverMock) GetSerialPortOutput(zone, name string) (string, error) {
	d.GetSerialPortOutputZone = zone
	d.GetSerialPortOutputName = name
	return d.GetSerialPortOutputResult, d.GetSerialPortOutputErr
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
