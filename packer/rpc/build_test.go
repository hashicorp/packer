package rpc

import (
	"errors"
	"github.com/mitchellh/packer/packer"
	"reflect"
	"testing"
)

var testBuildArtifact = &packer.MockArtifact{}

type testBuild struct {
	nameCalled      bool
	prepareCalled   bool
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

func (b *testBuild) Prepare() ([]string, error) {
	b.prepareCalled = true
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

func TestBuild(t *testing.T) {
	b := new(testBuild)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuild(b)
	bClient := client.Build()

	// Test Name
	bClient.Name()
	if !b.nameCalled {
		t.Fatal("name should be called")
	}

	// Test Prepare
	bClient.Prepare()
	if !b.prepareCalled {
		t.Fatal("prepare should be called")
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
	b := new(testBuild)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuild(b)
	bClient := client.Build()

	expected := []string{"foo"}
	b.prepareWarnings = expected

	warnings, err := bClient.Prepare()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if !reflect.DeepEqual(warnings, expected) {
		t.Fatalf("bad: %#v", warnings)
	}
}

func TestBuild_ImplementsBuild(t *testing.T) {
	var _ packer.Build = new(build)
}
