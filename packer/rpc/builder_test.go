package rpc

import (
	"cgl.tideland.biz/asserts"
	"errors"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

var testBuilderArtifact = &testArtifact{}

type testBuilder struct {
	prepareCalled bool
	prepareConfig []interface{}
	runCalled     bool
	runCache      packer.Cache
	runHook       packer.Hook
	runUi         packer.Ui
	cancelCalled  bool

	errRunResult bool
	nilRunResult bool
}

func (b *testBuilder) Prepare(config ...interface{}) error {
	b.prepareCalled = true
	b.prepareConfig = config
	return nil
}

func (b *testBuilder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	b.runCache = cache
	b.runCalled = true
	b.runHook = hook
	b.runUi = ui

	if b.errRunResult {
		return nil, errors.New("foo")
	} else if b.nilRunResult {
		return nil, nil
	} else {
		return testBuilderArtifact, nil
	}
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
	assert.Equal(b.prepareConfig, []interface{}{42}, "prepare should be called with right arg")

	// Test Run
	cache := new(testCache)
	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(ui, hook, cache)
	assert.Nil(err, "should have no error")
	assert.True(b.runCalled, "runs hould be called")

	if b.runCalled {
		b.runCache.Lock("foo")
		assert.True(cache.lockCalled, "lock should be called")

		b.runHook.Run("foo", nil, nil, nil)
		assert.True(hook.RunCalled, "run should be called")

		b.runUi.Say("format")
		assert.True(ui.sayCalled, "say should be called")
		assert.Equal(ui.sayMessage, "format", "message should be correct")

		assert.Equal(artifact.Id(), testBuilderArtifact.Id(), "should have artifact Id")
	}

	// Test run with nil result
	b.nilRunResult = true
	artifact, err = bClient.Run(ui, hook, cache)
	assert.Nil(artifact, "should be nil")
	assert.Nil(err, "should have no error")

	// Test with an error
	b.errRunResult = true
	b.nilRunResult = false
	artifact, err = bClient.Run(ui, hook, cache)
	assert.Nil(artifact, "should be nil")
	assert.NotNil(err, "should have error")

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
