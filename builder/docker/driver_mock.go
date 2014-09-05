package docker

import (
	"io"
)

// MockDriver is a driver implementation that can be used for tests.
type MockDriver struct {
	CommitCalled      bool
	CommitContainerId string
	CommitImageId     string
	CommitErr         error

	DeleteImageCalled bool
	DeleteImageId     string
	DeleteImageErr    error

	ImportCalled bool
	ImportPath   string
	ImportRepo   string
	ImportId     string
	ImportErr    error

	LoginCalled   bool
	LoginEmail    string
	LoginUsername string
	LoginPassword string
	LoginRepo     string
	LoginErr      error

	LogoutCalled bool
	LogoutRepo   string
	LogoutErr    error

	PushCalled bool
	PushName   string
	PushErr    error

	SaveImageCalled bool
	SaveImageId     string
	SaveImageReader io.Reader
	SaveImageError  error

	TagImageCalled  bool
	TagImageImageId string
	TagImageRepo    string
	TagImageErr     error

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

func (d *MockDriver) Commit(id string) (string, error) {
	d.CommitCalled = true
	d.CommitContainerId = id
	return d.CommitImageId, d.CommitErr
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

func (d *MockDriver) Login(r, e, u, p string) error {
	d.LoginCalled = true
	d.LoginRepo = r
	d.LoginEmail = e
	d.LoginUsername = u
	d.LoginPassword = p
	return d.LoginErr
}

func (d *MockDriver) Logout(r string) error {
	d.LogoutCalled = true
	d.LogoutRepo = r
	return d.LogoutErr
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

func (d *MockDriver) SaveImage(id string, dst io.Writer) error {
	d.SaveImageCalled = true
	d.SaveImageId = id

	if d.SaveImageReader != nil {
		_, err := io.Copy(dst, d.SaveImageReader)
		if err != nil {
			return err
		}
	}

	return d.SaveImageError
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

func (d *MockDriver) TagImage(id string, repo string) error {
	d.TagImageCalled = true
	d.TagImageImageId = id
	d.TagImageRepo = repo
	return d.TagImageErr
}

func (d *MockDriver) Verify() error {
	d.VerifyCalled = true
	return d.VerifyError
}
