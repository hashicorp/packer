package rpc

import (
	"cgl.tideland.biz/asserts"
	"errors"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

var testBuildArtifact = &testArtifact{}

type testBuild struct {
	nameCalled     bool
	prepareCalled  bool
	runCalled      bool
	runCache       packer.Cache
	runUi          packer.Ui
	setDebugCalled bool
	cancelCalled   bool

	errRunResult bool
}

func (b *testBuild) Name() string {
	b.nameCalled = true
	return "name"
}

func (b *testBuild) Prepare() error {
	b.prepareCalled = true
	return nil
}

func (b *testBuild) Run(ui packer.Ui, cache packer.Cache) (packer.Artifact, error) {
	b.runCalled = true
	b.runCache = cache
	b.runUi = ui

	if b.errRunResult {
		return nil, errors.New("foo")
	} else {
		return testBuildArtifact, nil
	}
}

func (b *testBuild) SetDebug(bool) {
	b.setDebugCalled = true
}

func (b *testBuild) Cancel() {
	b.cancelCalled = true
}

func TestBuildRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	b := new(testBuild)

	// Start the server
	server := rpc.NewServer()
	RegisterBuild(server, b)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")
	bClient := Build(client)

	// Test Name
	bClient.Name()
	assert.True(b.nameCalled, "name should be called")

	// Test Prepare
	bClient.Prepare()
	assert.True(b.prepareCalled, "prepare should be called")

	// Test Run
	cache := new(testCache)
	ui := new(testUi)
	_, err = bClient.Run(ui, cache)
	assert.True(b.runCalled, "run should be called")
	assert.Nil(err, "should not error")

	// Test the UI given to run, which should be fully functional
	if b.runCalled {
		b.runCache.Lock("foo")
		assert.True(cache.lockCalled, "lock should be called")

		b.runUi.Say("format")
		assert.True(ui.sayCalled, "say should be called")
		assert.Equal(ui.sayMessage, "format", "message should be correct")
	}

	// Test run with an error
	b.errRunResult = true
	_, err = bClient.Run(ui, cache)
	assert.NotNil(err, "should not nil")

	// Test SetDebug
	bClient.SetDebug(true)
	assert.True(b.setDebugCalled, "should be called")

	// Test Cancel
	bClient.Cancel()
	assert.True(b.cancelCalled, "cancel should be called")
}

func TestBuild_ImplementsBuild(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var realBuild packer.Build
	b := Build(nil)

	assert.Implementor(b, &realBuild, "should be a Build")
}
