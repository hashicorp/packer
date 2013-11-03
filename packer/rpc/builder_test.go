package rpc

import (
	"errors"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"reflect"
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

func (b *testBuilder) Prepare(config ...interface{}) ([]string, error) {
	b.prepareCalled = true
	b.prepareConfig = config
	return nil, nil
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
	// Create the interface to test
	b := new(testBuilder)

	// Start the server
	server := rpc.NewServer()
	RegisterBuilder(server, b)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test Prepare
	config := 42
	bClient := Builder(client)
	bClient.Prepare(config)
	if !b.prepareCalled {
		t.Fatal("should be called")
	}

	if !reflect.DeepEqual(b.prepareConfig, []interface{}{42}) {
		t.Fatalf("bad: %#v", b.prepareConfig)
	}

	// Test Run
	cache := new(testCache)
	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(ui, hook, cache)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !b.runCalled {
		t.Fatal("run should be called")
	}

	if b.runCalled {
		b.runCache.Lock("foo")
		if !cache.lockCalled {
			t.Fatal("should be called")
		}

		b.runHook.Run("foo", nil, nil, nil)
		if !hook.RunCalled {
			t.Fatal("should be called")
		}

		b.runUi.Say("format")
		if !ui.sayCalled {
			t.Fatal("say should be called")
		}

		if ui.sayMessage != "format" {
			t.Fatalf("bad: %s", ui.sayMessage)
		}

		if artifact.Id() != testBuilderArtifact.Id() {
			t.Fatalf("bad: %s", artifact.Id())
		}
	}

	// Test run with nil result
	b.nilRunResult = true
	artifact, err = bClient.Run(ui, hook, cache)
	if artifact != nil {
		t.Fatalf("bad: %#v", artifact)
	}
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}

	// Test with an error
	b.errRunResult = true
	b.nilRunResult = false
	artifact, err = bClient.Run(ui, hook, cache)
	if artifact != nil {
		t.Fatalf("bad: %#v", artifact)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test Cancel
	bClient.Cancel()
	if !b.cancelCalled {
		t.Fatal("cancel should be called")
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var _ packer.Builder = Builder(nil)
}
