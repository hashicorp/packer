package rpc

import (
	"github.com/mitchellh/packer/packer"
	"reflect"
	"testing"
)

func TestProvisionerRPC(t *testing.T) {
	// Create the interface to test
	p := new(packer.MockProvisioner)

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterProvisioner(p)
	pClient := client.Provisioner()

	// Test Prepare
	config := 42
	pClient.Prepare(config)
	if !p.PrepCalled {
		t.Fatal("should be called")
	}
	expected := []interface{}{int64(42)}
	if !reflect.DeepEqual(p.PrepConfigs, expected) {
		t.Fatalf("bad: %#v", p.PrepConfigs)
	}

	// Test Provision
	ui := &testUi{}
	comm := &packer.MockCommunicator{}
	pClient.Provision(ui, comm)
	if !p.ProvCalled {
		t.Fatal("should be called")
	}

	// Test Cancel
	pClient.Cancel()
	if !p.CancelCalled {
		t.Fatal("cancel should be called")
	}
}

func TestProvisioner_Implements(t *testing.T) {
	var _ packer.Provisioner = new(provisioner)
}
