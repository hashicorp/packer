package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

func TestHookRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the UI to test
	h := new(packer.MockHook)

	// Serve
	server := rpc.NewServer()
	RegisterHook(server, h)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")

	hClient := Hook(client)

	// Test Run
	ui := &testUi{}
	hClient.Run("foo", ui, nil, 42)
	assert.True(h.RunCalled, "run should be called")

	// Test Cancel
	hClient.Cancel()
	assert.True(h.CancelCalled, "cancel should be called")
}

func TestHook_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r packer.Hook
	h := &hook{nil}

	assert.Implementor(h, &r, "should be a Hook")
}
