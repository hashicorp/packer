package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

type testHook struct {
	runCalled     bool
	runUi         packer.Ui
}

func (h *testHook) Run(name string, data interface{}, ui packer.Ui) {
	h.runCalled = true
}

func TestHookRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the UI to test
	h := new(testHook)

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
	hClient.Run("foo", 42, ui)
	assert.True(h.runCalled, "run should be called")
}

func TestHook_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r packer.Hook
	h := &hook{nil}

	assert.Implementor(h, &r, "should be a Hook")
}
