package rpc

import (
	"github.com/mitchellh/packer/packer"
	"reflect"
	"testing"
)

var testEnvBuilder = &packer.MockBuilder{}
var testEnvCache = &testCache{}
var testEnvUi = &testUi{}

type testEnvironment struct {
	builderCalled bool
	builderName   string
	cliCalled     bool
	cliArgs       []string
	hookCalled    bool
	hookName      string
	ppCalled      bool
	ppName        string
	provCalled    bool
	provName      string
	uiCalled      bool
}

func (e *testEnvironment) Builder(name string) (packer.Builder, error) {
	e.builderCalled = true
	e.builderName = name
	return testEnvBuilder, nil
}

func (e *testEnvironment) Cache() packer.Cache {
	return testEnvCache
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

func (e *testEnvironment) PostProcessor(name string) (packer.PostProcessor, error) {
	e.ppCalled = true
	e.ppName = name
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
	// Create the interface to test
	e := &testEnvironment{}

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterEnvironment(e)
	eClient := client.Environment()

	// Test Builder
	builder, _ := eClient.Builder("foo")
	if !e.builderCalled {
		t.Fatal("builder should be called")
	}
	if e.builderName != "foo" {
		t.Fatalf("bad: %#v", e.builderName)
	}

	builder.Prepare(nil)
	if !testEnvBuilder.PrepareCalled {
		t.Fatal("should be called")
	}

	// Test Cache
	cache := eClient.Cache()
	cache.Lock("foo")
	if !testEnvCache.lockCalled {
		t.Fatal("should be called")
	}

	// Test Cli
	cliArgs := []string{"foo", "bar"}
	result, _ := eClient.Cli(cliArgs)
	if !e.cliCalled {
		t.Fatal("should be called")
	}
	if !reflect.DeepEqual(e.cliArgs, cliArgs) {
		t.Fatalf("bad: %#v", e.cliArgs)
	}
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Test Provisioner
	_, _ = eClient.Provisioner("foo")
	if !e.provCalled {
		t.Fatal("should be called")
	}
	if e.provName != "foo" {
		t.Fatalf("bad: %s", e.provName)
	}

	// Test Ui
	ui := eClient.Ui()
	if !e.uiCalled {
		t.Fatal("should be called")
	}

	// Test calls on the Ui
	ui.Say("format")
	if !testEnvUi.sayCalled {
		t.Fatal("should be called")
	}
	if testEnvUi.sayMessage != "format" {
		t.Fatalf("bad: %#v", testEnvUi.sayMessage)
	}
}

func TestEnvironment_ImplementsEnvironment(t *testing.T) {
	var _ packer.Environment = new(Environment)
}
