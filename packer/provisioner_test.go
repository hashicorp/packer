package packer

type TestProvisioner struct {
	prepCalled bool
	prepConfig interface{}
	prepUi     Ui
	provCalled bool
}

func (t *TestProvisioner) Prepare(config interface{}, ui Ui) {
	t.prepCalled = true
	t.prepConfig = config
	t.prepUi = ui
}

func (t *TestProvisioner) Provision(Ui, Communicator) {
	t.provCalled = true
}
