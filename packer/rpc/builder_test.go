package rpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
)

var testBuilderArtifact = &packer.MockArtifact{}

func TestBuilderPrepare(t *testing.T) {
	b := new(packer.MockBuilder)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	// Test Prepare
	config := 42
	warnings, err := bClient.Prepare(config)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}

	if !b.PrepareCalled {
		t.Fatal("should be called")
	}

	expected := []interface{}{int64(42)}
	if !reflect.DeepEqual(b.PrepareConfig, expected) {
		t.Fatalf("bad: %#v != %#v", b.PrepareConfig, expected)
	}
}

func TestBuilderPrepare_Warnings(t *testing.T) {
	b := new(packer.MockBuilder)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	expected := []string{"foo"}
	b.PrepareWarnings = expected

	// Test Prepare
	warnings, err := bClient.Prepare(nil)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if !reflect.DeepEqual(warnings, expected) {
		t.Fatalf("bad: %#v", warnings)
	}
}

func TestBuilderRun(t *testing.T) {
	b := new(packer.MockBuilder)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	// Test Run
	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(context.Background(), ui, hook)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !b.RunCalled {
		t.Fatal("run should be called")
	}

	if artifact.Id() != testBuilderArtifact.Id() {
		t.Fatalf("bad: %s", artifact.Id())
	}
}

func TestBuilderRun_nilResult(t *testing.T) {
	b := new(packer.MockBuilder)
	b.RunNilResult = true

	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(context.Background(), ui, hook)
	if artifact != nil {
		t.Fatalf("bad: %#v", artifact)
	}
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}
}

func TestBuilderRun_ErrResult(t *testing.T) {
	b := new(packer.MockBuilder)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	b.RunErrResult = true

	hook := &packer.MockHook{}
	ui := &testUi{}
	artifact, err := bClient.Run(context.Background(), ui, hook)
	if artifact != nil {
		t.Fatalf("bad: %#v", artifact)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderCancel(t *testing.T) {
	b := new(packer.MockBuilder)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	bClient.Run(ctx, nil, nil)

	if !b.CancelCalled {
		t.Fatal("cancel should be called")
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var _ packer.Builder = new(builder)
}
