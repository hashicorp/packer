package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

var testBuilderArtifact = &testArtifact{}

type testBuilder struct {
	prepareCalled bool
	prepareConfig interface{}
	runCalled     bool
	runHook       packer.Hook
	runUi         packer.Ui
	cancelCalled  bool
}

func (b *testBuilder) Prepare(config interface{}) error {
	b.prepareCalled = true
	b.prepareConfig = config
	return nil
}

func (b *testBuilder) Run(ui packer.Ui, hook packer.Hook) packer.Artifact {
	b.runCalled = true
	b.runHook = hook
	b.runUi = ui
	return testBuilderArtifact
}

func (b *testBuilder) Cancel() {
	b.cancelCalled = true
}

func TestBuilderRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	b := new(testBuilder)

	// Start the server
	server := rpc.NewServer()
	RegisterBuilder(server, b)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")

	// Test Prepare
	config := 42
	bClient := Builder(client)
	bClient.Prepare(config)
	assert.True(b.prepareCalled, "prepare should be called")
	assert.Equal(b.prepareConfig, 42, "prepare should be called with right arg")

	// Test Run
	hook := &testHook{}
	ui := &testUi{}
	artifact := bClient.Run(ui, hook)
	assert.True(b.runCalled, "runs hould be called")

	if b.runCalled {
		b.runHook.Run("foo", nil, nil, nil)
		assert.True(hook.runCalled, "run should be called")

		b.runUi.Say("format")
		assert.True(ui.sayCalled, "say should be called")
		assert.Equal(ui.sayMessage, "format", "message should be correct")

		assert.Equal(artifact.Id(), testBuilderArtifact.Id(), "should have artifact Id")
	}

	// Test Cancel
	bClient.Cancel()
	assert.True(b.cancelCalled, "cancel should be called")
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var realBuilder packer.Builder
	b := Builder(nil)

	assert.Implementor(b, &realBuilder, "should be a Builder")
}
