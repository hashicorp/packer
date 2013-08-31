package packer

import "testing"

func TestProvisionHook_Impl(t *testing.T) {
	var raw interface{}
	raw = &ProvisionHook{}
	if _, ok := raw.(Hook); !ok {
		t.Fatalf("must be a Hook")
	}
}

func TestProvisionHook(t *testing.T) {
	pA := &MockProvisioner{}
	pB := &MockProvisioner{}

	ui := testUi()
	var comm Communicator = nil
	var data interface{} = nil

	hook := &ProvisionHook{[]Provisioner{pA, pB}}
	hook.Run("foo", ui, comm, data)

	if !pA.ProvCalled {
		t.Error("provision should be called on pA")
	}

	if !pB.ProvCalled {
		t.Error("provision should be called on pB")
	}
}

// TODO(mitchellh): Test that they're run in the proper order
