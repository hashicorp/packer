package rpc

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

var testBuildArtifact = &packer.MockArtifact{}

type testBuild struct {
	nameCalled       bool
	prepareCalled    bool
	prepareWarnings  []string
	runFn            func(context.Context)
	runCalled        bool
	runUi            packersdk.Ui
	setDebugCalled   bool
	setForceCalled   bool
	setOnErrorCalled bool

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

func (b *testBuild) Run(ctx context.Context, ui packersdk.Ui) ([]packersdk.Artifact, error) {
	b.runCalled = true
	b.runUi = ui

	if b.runFn != nil {
		b.runFn(ctx)
	}

	if b.errRunResult {
		return nil, errors.New("foo")
	} else {
		return []packersdk.Artifact{testBuildArtifact}, nil
	}
}

func (b *testBuild) SetDebug(bool) {
	b.setDebugCalled = true
}

func (b *testBuild) SetForce(bool) {
	b.setForceCalled = true
}

func (b *testBuild) SetOnError(string) {
	b.setOnErrorCalled = true
}

func TestBuild(t *testing.T) {
	b := new(testBuild)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuild(b)
	bClient := client.Build()

	ctx := context.Background()

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
	ui := new(testUi)
	artifacts, err := bClient.Run(ctx, ui)
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
	_, err = bClient.Run(ctx, ui)
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

	// Test SetOnError
	bClient.SetOnError("ask")
	if !b.setOnErrorCalled {
		t.Fatal("should be called")
	}
}

func TestBuild_cancel(t *testing.T) {
	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	b := new(testBuild)

	done := make(chan interface{})
	b.runFn = func(ctx context.Context) {
		cancelTopCtx()
		<-ctx.Done()
		close(done)
	}

	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuild(b)
	bClient := client.Build()

	bClient.Prepare()

	ui := new(testUi)
	bClient.Run(topCtx, ui)

	// if context cancellation is not propagated, this will timeout
	<-done
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
