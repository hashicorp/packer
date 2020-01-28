package rpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
)

var testPostProcessorArtifact = new(packer.MockArtifact)

type TestPostProcessor struct {
	configCalled bool
	configVal    []interface{}
	ppCalled     bool
	ppArtifact   packer.Artifact
	ppArtifactId string
	ppUi         packer.Ui

	postProcessFn func(context.Context) error
}

func (*TestPostProcessor) ConfigSpec() hcldec.ObjectSpec { return nil }

func (pp *TestPostProcessor) Configure(v ...interface{}) error {
	pp.configCalled = true
	pp.configVal = v
	return nil
}

func (pp *TestPostProcessor) PostProcess(ctx context.Context, ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, bool, error) {
	pp.ppCalled = true
	pp.ppArtifact = a
	pp.ppArtifactId = a.Id()
	pp.ppUi = ui
	if pp.postProcessFn != nil {
		return testPostProcessorArtifact, false, false, pp.postProcessFn(ctx)
	}
	return testPostProcessorArtifact, false, false, nil
}

func TestPostProcessorRPC(t *testing.T) {
	// Create the interface to test
	p := new(TestPostProcessor)

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterPostProcessor(p)

	ppClient := client.PostProcessor()

	// Test Configure
	config := 42
	err := ppClient.Configure(config)
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	if !p.configCalled {
		t.Fatal("config should be called")
	}

	expected := []interface{}{int64(42)}
	if !reflect.DeepEqual(p.configVal, expected) {
		t.Fatalf("unknown config value: %#v", p.configVal)
	}

	// Test PostProcess
	a := &packer.MockArtifact{
		IdValue: "ppTestId",
	}
	ui := new(testUi)
	artifact, _, _, err := ppClient.PostProcess(context.Background(), ui, a)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !p.ppCalled {
		t.Fatal("postprocess should be called")
	}

	if p.ppArtifactId != "ppTestId" {
		t.Fatalf("unknown artifact: '%s'", p.ppArtifact.Id())
	}

	if artifact.Id() != "id" {
		t.Fatalf("unknown artifact: %s", artifact.Id())
	}
}

func TestPostProcessorRPC_cancel(t *testing.T) {
	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	p := new(TestPostProcessor)
	p.postProcessFn = func(ctx context.Context) error {
		cancelTopCtx()
		<-ctx.Done()
		return ctx.Err()
	}

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	if err := server.RegisterPostProcessor(p); err != nil {
		panic(err)
	}

	ppClient := client.PostProcessor()

	// Test Configure
	config := 42
	err := ppClient.Configure(config)
	if err != nil {
		t.Fatalf("error configuring post-processor client: %s", err)
	}

	// Test PostProcess
	a := &packer.MockArtifact{
		IdValue: "ppTestId",
	}
	ui := new(testUi)
	_, _, _, err = ppClient.PostProcess(topCtx, ui, a)
	if err == nil {
		t.Fatalf("should err")
	}
}

func TestPostProcessor_Implements(t *testing.T) {
	var raw interface{}
	raw = new(postProcessor)
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatal("not a postprocessor")
	}
}
