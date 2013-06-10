package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

var testBuildArtifact = &testArtifact{}

type testBuild struct {
	nameCalled    bool
	prepareCalled bool
	prepareUi     packer.Ui
	runCalled     bool
	runCache      packer.Cache
	runUi         packer.Ui
	cancelCalled  bool
}

func (b *testBuild) Name() string {
	b.nameCalled = true
	return "name"
}

func (b *testBuild) Prepare(ui packer.Ui) error {
	b.prepareCalled = true
	b.prepareUi = ui
	return nil
}

func (b *testBuild) Run(ui packer.Ui, cache packer.Cache) packer.Artifact {
	b.runCalled = true
	b.runCache = cache
	b.runUi = ui
	return testBuildArtifact
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
	ui := new(testUi)
	bClient.Prepare(ui)
	assert.True(b.prepareCalled, "prepare should be called")

	// Test Run
	cache := new(testCache)
	ui = new(testUi)
	bClient.Run(ui, cache)
	assert.True(b.runCalled, "run should be called")

	// Test the UI given to run, which should be fully functional
	if b.runCalled {
		b.runCache.Lock("foo")
		assert.True(cache.lockCalled, "lock should be called")

		b.runUi.Say("format")
		assert.True(ui.sayCalled, "say should be called")
		assert.Equal(ui.sayMessage, "format", "message should be correct")
	}

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
