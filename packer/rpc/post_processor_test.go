package rpc

import (
	"github.com/mitchellh/packer/packer"
	"reflect"
	"testing"
)

var testPostProcessorArtifact = new(packer.MockArtifact)

type TestPostProcessor struct {
	configCalled bool
	configVal    []interface{}
	ppCalled     bool
	ppArtifact   packer.Artifact
	ppUi         packer.Ui
}

func (pp *TestPostProcessor) Configure(v ...interface{}) error {
	pp.configCalled = true
	pp.configVal = v
	return nil
}

func (pp *TestPostProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	pp.ppCalled = true
	pp.ppArtifact = a
	pp.ppUi = ui
	return testPostProcessorArtifact, false, nil
}

func TestPostProcessorRPC(t *testing.T) {
	// Create the interface to test
	p := new(TestPostProcessor)

	// Start the server
	server := NewServer()
	server.RegisterPostProcessor(p)
	client := testClient(t, server)
	defer client.Close()
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

	if !reflect.DeepEqual(p.configVal, []interface{}{42}) {
		t.Fatalf("unknown config value: %#v", p.configVal)
	}

	// Test PostProcess
	a := &packer.MockArtifact{
		IdValue: "ppTestId",
	}
	ui := new(testUi)
	artifact, _, err := ppClient.PostProcess(ui, a)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !p.ppCalled {
		t.Fatal("postprocess should be called")
	}

	if p.ppArtifact.Id() != "ppTestId" {
		t.Fatal("unknown artifact")
	}

	if artifact.Id() != "id" {
		t.Fatalf("unknown artifact: %s", artifact.Id())
	}
}

func TestPostProcessor_Implements(t *testing.T) {
	var raw interface{}
	raw = PostProcessor(nil)
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatal("not a postprocessor")
	}
}
