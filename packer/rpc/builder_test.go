package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"reflect"
	"testing"
)

var testBuilderArtifact = &testArtifact{}

func builderRPCClient(t *testing.T) (*packer.MockBuilder, packer.Builder) {
	b := new(packer.MockBuilder)

	// Start the server
	server := rpc.NewServer()
	RegisterBuilder(server, b)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return b, Builder(client)
}

func TestBuilderPrepare(t *testing.T) {
	b, bClient := builderRPCClient(t)

	// Test Prepare
	config := 42
	bClient.Prepare(config)
	if !b.PrepareCalled {
		t.Fatal("should be called")
	}

	if !reflect.DeepEqual(b.PrepareConfig, []interface{}{42}) {
		t.Fatalf("bad: %#v", b.PrepareConfig)
	}
}

func TestBuilderRun(t *testing.T) {
	b, bClient := builderRPCClient(t)

	// Test Run
	cache := new(testCache)
	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(ui, hook, cache)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !b.RunCalled {
		t.Fatal("run should be called")
	}

	b.RunCache.Lock("foo")
	if !cache.lockCalled {
		t.Fatal("should be called")
	}

	b.RunHook.Run("foo", nil, nil, nil)
	if !hook.RunCalled {
		t.Fatal("should be called")
	}

	b.RunUi.Say("format")
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

func TestBuilderRun_nilResult(t *testing.T) {
	b, bClient := builderRPCClient(t)
	b.RunNilResult = true

	cache := new(testCache)
	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(ui, hook, cache)
	if artifact != nil {
		t.Fatalf("bad: %#v", artifact)
	}
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}
}

func TestBuilderRun_ErrResult(t *testing.T) {
	b, bClient := builderRPCClient(t)
	b.RunErrResult = true

	cache := new(testCache)
	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(ui, hook, cache)
	if artifact != nil {
		t.Fatalf("bad: %#v", artifact)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderCancel(t *testing.T) {
	b, bClient := builderRPCClient(t)

	bClient.Cancel()
	if !b.CancelCalled {
		t.Fatal("cancel should be called")
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var _ packer.Builder = Builder(nil)
}
