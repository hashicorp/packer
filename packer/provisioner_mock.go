package packer

// MockProvisioner is an implementation of Provisioner that can be
// used for tests.
type MockProvisioner struct {
	PrepCalled   bool
	PrepConfigs  []interface{}
	ProvCalled   bool
	ProvUi       Ui
	CancelCalled bool
}

func (t *MockProvisioner) Prepare(configs ...interface{}) error {
	t.PrepCalled = true
	t.PrepConfigs = configs
	return nil
}

func (t *MockProvisioner) Provision(ui Ui, comm Communicator) error {
	t.ProvCalled = true
	t.ProvUi = ui
	return nil
}

func (t *MockProvisioner) Cancel() {
	t.CancelCalled = true
}
