package digitalocean

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactId(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, "San Francisco", nil}
	expected := "42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, "San Francisco", nil}
	expected := "A snapshot was created: 'packer-foobar' in region 'San Francisco'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
