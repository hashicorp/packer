package docker

import (
	"io"
)

// MockDriver is a driver implementation that can be used for tests.
type MockDriver struct {
	DeleteImageCalled bool
	DeleteImageId     string
	DeleteImageErr    error

	ImportCalled bool
	ImportPath   string
	ImportRepo   string
	ImportId     string
	ImportErr    error

	PushCalled bool
	PushName   string
	PushErr    error

	ExportReader io.Reader
	ExportError  error
	PullError    error
	StartID      string
	StartError   error
	StopError    error
	VerifyError  error

	ExportCalled bool
	ExportID     string
	PullCalled   bool
	PullImage    string
	StartCalled  bool
	StartConfig  *ContainerConfig
	StopCalled   bool
	StopID       string
	VerifyCalled bool
}

func (d *MockDriver) DeleteImage(id string) error {
	d.DeleteImageCalled = true
	d.DeleteImageId = id
	return d.DeleteImageErr
}

func (d *MockDriver) Export(id string, dst io.Writer) error {
	d.ExportCalled = true
	d.ExportID = id

	if d.ExportReader != nil {
		_, err := io.Copy(dst, d.ExportReader)
		if err != nil {
			return err
		}
	}

	return d.ExportError
}

func (d *MockDriver) Import(path, repo string) (string, error) {
	d.ImportCalled = true
	d.ImportPath = path
	d.ImportRepo = repo
	return d.ImportId, d.ImportErr
}

func (d *MockDriver) Pull(image string) error {
	d.PullCalled = true
	d.PullImage = image
	return d.PullError
}

func (d *MockDriver) Push(name string) error {
	d.PushCalled = true
	d.PushName = name
	return d.PushErr
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

func (d *MockDriver) Verify() error {
	d.VerifyCalled = true
	return d.VerifyError
}
