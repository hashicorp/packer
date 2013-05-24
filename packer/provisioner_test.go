package packer

import "testing"

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

func TestProvisionHook_Impl(t *testing.T) {
	var raw interface{}
	raw = &ProvisionHook{}
	if _, ok := raw.(Hook); !ok {
		t.Fatalf("must be a Hook")
	}
}

func TestProvisionHook(t *testing.T) {
	pA := &TestProvisioner{}
	pB := &TestProvisioner{}

	ui := testUi()
	var comm Communicator = nil
	var data interface{} = nil

	hook := &ProvisionHook{[]Provisioner{pA, pB}}
	hook.Run("foo", ui, comm, data)

	if !pA.provCalled {
		t.Error("provision should be called on pA")
	}

	if !pB.provCalled {
		t.Error("provision should be called on pB")
	}
}

// TODO(mitchellh): Test that they're run in the proper order
