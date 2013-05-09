package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

type testBuilder struct {
	prepareCalled bool
	prepareConfig interface{}
	runCalled bool
	runBuild packer.Build
	runUi packer.Ui
}

func (b *testBuilder) Prepare(config interface{}) error {
	b.prepareCalled = true
	b.prepareConfig = config
	return nil
}

func (b *testBuilder) Run(build packer.Build, ui packer.Ui) {
	b.runCalled = true
	b.runBuild = build
	b.runUi = ui
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
	build := &testBuild{}
	ui := &testUi{}
	bClient.Run(build, ui)
	assert.True(b.runCalled, "runs hould be called")

	if b.runCalled {
		b.runBuild.Prepare()
		assert.True(build.prepareCalled, "prepare should be called")

		b.runUi.Say("format")
		assert.True(ui.sayCalled, "say should be called")
		assert.Equal(ui.sayFormat, "format", "format should be correct")
	}
}

func TestBuilder_ImplementsBuild(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var realBuilder packer.Builder
	b := Builder(nil)

	assert.Implementor(b, &realBuilder, "should be a Builder")
}
