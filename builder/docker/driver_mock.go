package docker

// MockDriver is a driver implementation that can be used for tests.
type MockDriver struct {
	PullError error

	PullCalled bool
	PullImage  string
}

func (d *MockDriver) Pull(image string) error {
	d.PullCalled = true
	d.PullImage = image
	return d.PullError
}
