package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

type testBuild struct {
	nameCalled bool
	prepareCalled bool
	runCalled bool
	runUi packer.Ui
}

func (b *testBuild) Name() string {
	b.nameCalled = true
	return "name"
}

func (b *testBuild) Prepare() error {
	b.prepareCalled = true
	return nil
}

func (b *testBuild) Run(ui packer.Ui) {
	b.runCalled = true
	b.runUi = ui
}

func TestBuildRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the UI to test
	b := new(testBuild)
	bServer := &BuildServer{b}

	// Start the RPC server
	readyChan := make(chan int)
	stopChan := make(chan int)
	defer func() { stopChan <- 1 }()
	go testRPCServer(":1234", "Build", bServer, readyChan, stopChan)
	<-readyChan

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		panic(err)
	}

	bClient := &Build{client}

	// Test Name
	bClient.Name()
	assert.True(b.nameCalled, "name should be called")

	// Test Prepare
	bClient.Prepare()
	assert.True(b.prepareCalled, "prepare should be called")

	// Test Run
	ui := new(testUi)
	bClient.Run(ui)
	assert.True(b.runCalled, "run should be called")

	// Test the UI given to run, which should be fully functional
	if b.runCalled {
		b.runUi.Say("format")
		assert.True(ui.sayCalled, "say should be called")
		assert.Equal(ui.sayFormat, "format", "format should be correct")
	}
}

func TestBuild_ImplementsBuild(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var realBuild packer.Build
	b := &Build{nil}

	assert.Implementor(b, &realBuild, "should be a Build")
}
