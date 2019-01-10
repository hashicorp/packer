package rpc

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
)

type TestPreProcessor struct {
	configCalled bool
	configVal    []interface{}
	ppCalled     bool
	ppUi         packer.Ui
}

func (pp *TestPreProcessor) Configure(v ...interface{}) error {
	pp.configCalled = true
	pp.configVal = v
	return nil
}

func (pp *TestPreProcessor) PreProcess(ui packer.Ui) error {
	pp.ppCalled = true
	pp.ppUi = ui
	return nil
}

func TestPreProcessorRPC(t *testing.T) {
	// Create the interface to test
	p := new(TestPreProcessor)

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterPreProcessor(p)

	ppClient := client.PreProcessor()

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

	// Test PreProcess
	ui := new(testUi)
	err = ppClient.PreProcess(ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !p.ppCalled {
		t.Fatal("preprocess should be called")
	}
}

func TestPreProcessor_Implements(t *testing.T) {
	var raw interface{}
	raw = new(preProcessor)
	if _, ok := raw.(packer.PreProcessor); !ok {
		t.Fatal("not a preprocessor")
	}
}
