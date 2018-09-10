package packer

// MockProvisioner is an implementation of Provisioner that can be
// used for tests.
type MockProvisioner struct {
	ProvFunc func() error

	PrepCalled       bool
	PrepConfigs      []interface{}
	ProvCalled       bool
	ProvCommunicator Communicator
	ProvUi           Ui
	CancelCalled     bool
}

func (t *MockProvisioner) Prepare(configs ...interface{}) error {
	t.PrepCalled = true
	t.PrepConfigs = configs
	return nil
}

func (t *MockProvisioner) Provision(ui Ui, comm Communicator) error {
	t.ProvCalled = true
	t.ProvCommunicator = comm
	t.ProvUi = ui

	if t.ProvFunc == nil {
		return nil
	}

	return t.ProvFunc()
}

func (t *MockProvisioner) Cancel() {
	t.CancelCalled = true
}
