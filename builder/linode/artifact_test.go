package linode

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
	a := &Artifact{"private/42", "packer-foobar", nil}
	expected := "private/42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"private/42", "packer-foobar", nil}
	expected := "Linode image: packer-foobar (private/42)"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
