package packer

type TestProvisioner struct {
	prepCalled bool
	provCalled bool
}

func (t *TestProvisioner) Prepare(interface{}, Ui) {
	t.prepCalled = true
}

func (t *TestProvisioner) Provision(Ui, Communicator) {
	t.provCalled = true
}
