package rpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

var testBuilderArtifact = &packersdk.MockArtifact{}

func TestBuilderPrepare(t *testing.T) {
	b := new(packer.MockBuilder)
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	// Test Prepare
	config := 42
	_, warnings, err := bClient.Prepare(config)
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
	_, warnings, err := bClient.Prepare(nil)
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
	hook := &packersdk.MockHook{}
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

	hook := &packersdk.MockHook{}
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

	hook := &packersdk.MockHook{}
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
	topCtx, topCtxCancel := context.WithCancel(context.Background())
	// var runCtx context.Context

	b := new(packer.MockBuilder)
	cancelled := false
	b.RunFn = func(ctx context.Context) {
		topCtxCancel()
		<-ctx.Done()
		cancelled = true
	}
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterBuilder(b)
	bClient := client.Builder()

	_, err := bClient.Run(topCtx, new(testUi), new(packersdk.MockHook))
	if err != nil {
		t.Fatalf("mock shouldnt retun run error for cancellation")
	}

	if !cancelled {
		t.Fatal("context should have been cancelled")
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var _ packer.Builder = new(builder)
}
