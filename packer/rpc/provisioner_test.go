package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

type testProvisioner struct {
	prepareCalled  bool
	prepareConfigs []interface{}
	provCalled     bool
	provComm       packer.Communicator
	provUi         packer.Ui
}

func (p *testProvisioner) Prepare(configs ...interface{}) error {
	p.prepareCalled = true
	p.prepareConfigs = configs
	return nil
}

func (p *testProvisioner) Provision(ui packer.Ui, comm packer.Communicator) {
	p.provCalled = true
	p.provComm = comm
	p.provUi = ui
}

func TestProvisionerRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	p := new(testProvisioner)

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
	assert.True(p.prepareCalled, "prepare should be called")
	assert.Equal(p.prepareConfigs, []interface{}{42}, "prepare should be called with right arg")

	// Test Provision
	ui := &testUi{}
	comm := &testCommunicator{}
	pClient.Provision(ui, comm)
	assert.True(p.provCalled, "provision should be called")

	p.provUi.Say("foo")
	assert.True(ui.sayCalled, "say should be called")
}

func TestProvisioner_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r packer.Provisioner
	p := Provisioner(nil)

	assert.Implementor(p, &r, "should be a provisioner")
}
