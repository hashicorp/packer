package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"reflect"
	"testing"
)

func TestProvisionerRPC(t *testing.T) {
	// Create the interface to test
	p := new(packer.MockProvisioner)

	// Start the server
	server := rpc.NewServer()
	RegisterProvisioner(server, p)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test Prepare
	config := 42
	pClient := Provisioner(client)
	pClient.Prepare(config)
	if !p.PrepCalled {
		t.Fatal("should be called")
	}
	if !reflect.DeepEqual(p.PrepConfigs, []interface{}{42}) {
		t.Fatalf("bad: %#v", p.PrepConfigs)
	}

	// Test Provision
	ui := &testUi{}
	comm := &packer.MockCommunicator{}
	pClient.Provision(ui, comm)
	if !p.ProvCalled {
		t.Fatal("should be called")
	}

	p.ProvUi.Say("foo")
	if !ui.sayCalled {
		t.Fatal("should be called")
	}

	// Test Cancel
	pClient.Cancel()
	if !p.CancelCalled {
		t.Fatal("cancel should be called")
	}
}

func TestProvisioner_Implements(t *testing.T) {
	var _ packer.Provisioner = Provisioner(nil)
}
