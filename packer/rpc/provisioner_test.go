package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

func TestProvisionerRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	p := new(packer.MockProvisioner)

	// Start the server
	server := rpc.NewServer()
	RegisterProvisioner(server, p)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")

	// Test Prepare
	config := 42
	pClient := Provisioner(client)
	pClient.Prepare(config)
	assert.True(p.PrepCalled, "prepare should be called")
	assert.Equal(p.PrepConfigs, []interface{}{42}, "prepare should be called with right arg")

	// Test Provision
	ui := &testUi{}
	comm := &packer.MockCommunicator{}
	pClient.Provision(ui, comm)
	assert.True(p.ProvCalled, "provision should be called")

	p.ProvUi.Say("foo")
	assert.True(ui.sayCalled, "say should be called")

	// Test Cancel
	pClient.Cancel()
	if !p.CancelCalled {
		t.Fatal("cancel should be called")
	}
}

func TestProvisioner_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r packer.Provisioner
	p := Provisioner(nil)

	assert.Implementor(p, &r, "should be a provisioner")
}
