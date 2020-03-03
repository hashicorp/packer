package vagrant

// Create a mock driver so that we can test Vagrant builder steps
type MockVagrantDriver struct {
	InitCalled      bool
	AddCalled       bool
	UpCalled        bool
	HaltCalled      bool
	SuspendCalled   bool
	SSHConfigCalled bool
	DestroyCalled   bool
	PackageCalled   bool
	VerifyCalled    bool
	VersionCalled   bool

	ReturnError     error
	ReturnSSHConfig *VagrantSSHConfig
	GlobalID        string
}

func (d *MockVagrantDriver) Init([]string) error {
	d.InitCalled = true
	return d.ReturnError
}

func (d *MockVagrantDriver) Add([]string) error {
	d.AddCalled = true
	return d.ReturnError
}

func (d *MockVagrantDriver) Up([]string) (string, string, error) {
	d.UpCalled = true
	return "", "", nil
}

func (d *MockVagrantDriver) Halt(string) error {
	d.HaltCalled = true
	return d.ReturnError
}

func (d *MockVagrantDriver) Suspend(string) error {
	d.SuspendCalled = true
	return d.ReturnError
}

func (d *MockVagrantDriver) SSHConfig(gid string) (*VagrantSSHConfig, error) {
	d.SSHConfigCalled = true
	// track the input value
	d.GlobalID = gid

	if d.ReturnSSHConfig != nil {
		return d.ReturnSSHConfig, nil
	}

	sshConfig := VagrantSSHConfig{
		Hostname:               "127.0.0.1",
		User:                   "vagrant",
		Port:                   "2222",
		UserKnownHostsFile:     "/dev/null",
		StrictHostKeyChecking:  false,
		PasswordAuthentication: false,
		IdentityFile:           "\"/path with spaces/insecure_private_key\"",
		IdentitiesOnly:         true,
		LogLevel:               "FATAL"}
	return &sshConfig, d.ReturnError
}

func (d *MockVagrantDriver) Destroy(string) error {
	d.DestroyCalled = true
	return d.ReturnError
}

func (d *MockVagrantDriver) Package([]string) error {
	d.PackageCalled = true
	return d.ReturnError
}

func (d *MockVagrantDriver) Verify() error {
	d.VerifyCalled = true
	return d.ReturnError
}

func (d *MockVagrantDriver) Version() (string, error) {
	d.VersionCalled = true
	return "", d.ReturnError
}

// End of mock definition
