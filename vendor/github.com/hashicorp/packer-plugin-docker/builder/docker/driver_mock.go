package docker

import (
	"io"

	"github.com/hashicorp/go-version"
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

	IPAddressCalled bool
	IPAddressID     string
	IPAddressResult string
	IPAddressErr    error

	Sha256Called bool
	Sha256Id     string
	Sha256Result string
	Sha256Err    error

	KillCalled bool
	KillID     string
	KillError  error

	LoginCalled   bool
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

	TagImageCalled  int
	TagImageImageId string
	TagImageRepo    []string
	TagImageForce   bool
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

	VersionCalled  bool
	VersionVersion string
}

func (d *MockDriver) Commit(id string, author string, changes []string, message string) (string, error) {
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

func (d *MockDriver) Import(path string, changes []string, repo string) (string, error) {
	d.ImportCalled = true
	d.ImportPath = path
	d.ImportRepo = repo
	return d.ImportId, d.ImportErr
}

func (d *MockDriver) IPAddress(id string) (string, error) {
	d.IPAddressCalled = true
	d.IPAddressID = id
	return d.IPAddressResult, d.IPAddressErr
}

func (d *MockDriver) Sha256(id string) (string, error) {
	d.Sha256Called = true
	d.Sha256Id = id
	return d.Sha256Result, d.Sha256Err
}

func (d *MockDriver) Login(r, u, p string) error {
	d.LoginCalled = true
	d.LoginRepo = r
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

func (d *MockDriver) KillContainer(id string) error {
	d.KillCalled = true
	d.KillID = id
	return d.KillError
}

func (d *MockDriver) StopContainer(id string) error {
	d.StopCalled = true
	d.StopID = id
	return d.StopError
}

func (d *MockDriver) TagImage(id string, repo string, force bool) error {
	d.TagImageCalled += 1
	d.TagImageImageId = id
	d.TagImageRepo = append(d.TagImageRepo, repo)
	d.TagImageForce = force
	return d.TagImageErr
}

func (d *MockDriver) Verify() error {
	d.VerifyCalled = true
	return d.VerifyError
}

func (d *MockDriver) Version() (*version.Version, error) {
	d.VersionCalled = true
	return version.NewVersion(d.VersionVersion)
}
