package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

var testEnvBuilder = &testBuilder{}
var testEnvUi = &testUi{}

type testEnvironment struct {
	builderCalled bool
	builderName   string
	cliCalled     bool
	cliArgs       []string
	hookCalled    bool
	hookName      string
	provCalled    bool
	provName      string
	uiCalled      bool
}

func (e *testEnvironment) Builder(name string) (packer.Builder, error) {
	e.builderCalled = true
	e.builderName = name
	return testEnvBuilder, nil
}

func (e *testEnvironment) Cli(args []string) (int, error) {
	e.cliCalled = true
	e.cliArgs = args
	return 42, nil
}

func (e *testEnvironment) Hook(name string) (packer.Hook, error) {
	e.hookCalled = true
	e.hookName = name
	return nil, nil
}

func (e *testEnvironment) Provisioner(name string) (packer.Provisioner, error) {
	e.provCalled = true
	e.provName = name
	return nil, nil
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
	server := rpc.NewServer()
	RegisterEnvironment(server, e)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")
	eClient := &Environment{client}

	// Test Builder
	builder, _ := eClient.Builder("foo")
	assert.True(e.builderCalled, "Builder should be called")
	assert.Equal(e.builderName, "foo", "Correct name for Builder")

	builder.Prepare(nil)
	assert.True(testEnvBuilder.prepareCalled, "Prepare should be called")

	// Test Cli
	cliArgs := []string{"foo", "bar"}
	result, _ := eClient.Cli(cliArgs)
	assert.True(e.cliCalled, "CLI should be called")
	assert.Equal(e.cliArgs, cliArgs, "args should match")
	assert.Equal(result, 42, "result shuld be 42")

	// Test Provisioner
	_, _ = eClient.Provisioner("foo")
	assert.True(e.provCalled, "provisioner should be called")
	assert.Equal(e.provName, "foo", "should have proper name")

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
