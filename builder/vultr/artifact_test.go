package vultr

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactId(t *testing.T) {
	a := &Artifact{"d455d0246e8e6", "packer-test", nil}
	expected := "d455d0246e8e6"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"d455d0246e8e6", "packer-test", nil}
	expected := "Vultr Snapshot: packer-test (d455d0246e8e6)"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
