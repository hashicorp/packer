package docker

// MockDriver is a driver implementation that can be used for tests.
type MockDriver struct {
	PullError  error
	StartID    string
	StartError error
	StopError  error

	PullCalled  bool
	PullImage   string
	StartCalled bool
	StartConfig *ContainerConfig
	StopCalled  bool
	StopID      string
}

func (d *MockDriver) Pull(image string) error {
	d.PullCalled = true
	d.PullImage = image
	return d.PullError
}

func (d *MockDriver) StartContainer(config *ContainerConfig) (string, error) {
	d.StartCalled = true
	d.StartConfig = config
	return d.StartID, d.StartError
}

func (d *MockDriver) StopContainer(id string) error {
	d.StopCalled = true
	d.StopID = id
	return d.StopError
}
