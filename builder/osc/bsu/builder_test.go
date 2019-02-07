package bsu

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_key":   "foo",
		"secret_key":   "bar",
		"source_omi":   "foo",
		"vm_type":      "foo",
		"region":       "us-east-1",
		"ssh_username": "root",
		"omi_name":     "foo",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}
