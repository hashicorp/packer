package rpc

import (
	"errors"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"reflect"
	"testing"
)

var testBuildArtifact = &testArtifact{}

type testBuild struct {
	nameCalled      bool
	prepareCalled   bool
	prepareVars     map[string]string
	prepareWarnings []string
	runCalled       bool
	runCache        packer.Cache
	runUi           packer.Ui
	setDebugCalled  bool
	setForceCalled  bool
	cancelCalled    bool

	errRunResult bool
}

func (b *testBuild) Name() string {
	b.nameCalled = true
	return "name"
}

func (b *testBuild) Prepare(v map[string]string) ([]string, error) {
	b.prepareCalled = true
	b.prepareVars = v
	return b.prepareWarnings, nil
}

func (b *testBuild) Run(ui packer.Ui, cache packer.Cache) ([]packer.Artifact, error) {
	b.runCalled = true
	b.runCache = cache
	b.runUi = ui

	if b.errRunResult {
		return nil, errors.New("foo")
	} else {
		return []packer.Artifact{testBuildArtifact}, nil
	}
}

func (b *testBuild) SetDebug(bool) {
	b.setDebugCalled = true
}

func (b *testBuild) SetForce(bool) {
	b.setForceCalled = true
}

func (b *testBuild) Cancel() {
	b.cancelCalled = true
}

func buildRPCClient(t *testing.T) (*testBuild, packer.Build) {
	// Create the interface to test
	b := new(testBuild)

	// Start the server
	server := rpc.NewServer()
	RegisterBuild(server, b)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return b, Build(client)
}

func TestBuild(t *testing.T) {
	b, bClient := buildRPCClient(t)

	// Test Name
	bClient.Name()
	if !b.nameCalled {
		t.Fatal("name should be called")
	}

	// Test Prepare
	bClient.Prepare(map[string]string{"foo": "bar"})
	if !b.prepareCalled {
		t.Fatal("prepare should be called")
	}

	if len(b.prepareVars) != 1 {
		t.Fatalf("bad vars: %#v", b.prepareVars)
	}

	if b.prepareVars["foo"] != "bar" {
		t.Fatalf("bad vars: %#v", b.prepareVars)
	}

	// Test Run
	cache := new(testCache)
	ui := new(testUi)
	artifacts, err := bClient.Run(ui, cache)
	if !b.runCalled {
		t.Fatal("run should be called")
	}

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(artifacts) != 1 {
		t.Fatalf("bad: %#v", artifacts)
	}

	if artifacts[0].BuilderId() != "bid" {
		t.Fatalf("bad: %#v", artifacts)
	}

	// Test the UI given to run, which should be fully functional
	if b.runCalled {
		b.runCache.Lock("foo")
		if !cache.lockCalled {
			t.Fatal("lock shuld be called")
		}

		b.runUi.Say("format")
		if !ui.sayCalled {
			t.Fatal("say should be called")
		}

		if ui.sayMessage != "format" {
			t.Fatalf("bad: %#v", ui.sayMessage)
		}
	}

	// Test run with an error
	b.errRunResult = true
	_, err = bClient.Run(ui, cache)
	if err == nil {
		t.Fatal("should error")
	}

	// Test SetDebug
	bClient.SetDebug(true)
	if !b.setDebugCalled {
		t.Fatal("should be called")
	}

	// Test SetForce
	bClient.SetForce(true)
	if !b.setForceCalled {
		t.Fatal("should be called")
	}

	// Test Cancel
	bClient.Cancel()
	if !b.cancelCalled {
		t.Fatal("should be called")
	}
}

func TestBuildPrepare_Warnings(t *testing.T) {
	b, bClient := buildRPCClient(t)

	expected := []string{"foo"}
	b.prepareWarnings = expected

	warnings, err := bClient.Prepare(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if !reflect.DeepEqual(warnings, expected) {
		t.Fatalf("bad: %#v", warnings)
	}
}

func TestBuild_ImplementsBuild(t *testing.T) {
	var _ packer.Build = Build(nil)
}
