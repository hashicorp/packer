package brkt

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should implement the packer.Artifact interface")
	}
}
