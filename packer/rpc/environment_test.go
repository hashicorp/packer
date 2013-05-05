package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

var testEnvUi = &testUi{}

type testEnvironment struct {
	bfCalled bool
	cliCalled bool
	cliArgs []string
	uiCalled bool
}

func (e *testEnvironment) Cli(args []string) int {
	e.cliCalled = true
	e.cliArgs = args
	return 42
}

func (e *testEnvironment) Ui() packer.Ui {
	e.uiCalled = true
	return testEnvUi
}

func TestEnvironmentRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	e := &testEnvironment{}

	// Start the server
	server := NewServer()
	server.RegisterEnvironment(e)
	server.Start()
	defer server.Stop()

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", server.Address())
	assert.Nil(err, "should be able to connect")
	eClient := &Environment{client}

	// Test Cli
	cliArgs := []string{"foo", "bar"}
	result := eClient.Cli(cliArgs)
	assert.True(e.cliCalled, "CLI should be called")
	assert.Equal(e.cliArgs, cliArgs, "args should match")
	assert.Equal(result, 42, "result shuld be 42")

	// Test Ui
	ui := eClient.Ui()
	assert.True(e.uiCalled, "Ui should've been called")

	// Test calls on the Ui
	ui.Say("format")
	assert.True(testEnvUi.sayCalled, "Say should be called")
	assert.Equal(testEnvUi.sayFormat, "format", "format should match")
}

func TestEnvironment_ImplementsEnvironment(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var realVar packer.Environment
	e := &Environment{nil}

	assert.Implementor(e, &realVar, "should be an Environment")
}
