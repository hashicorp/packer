package digitalocean

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, "San Francisco", nil}
	expected := "A snapshot was created: 'packer-foobar' in region 'San Francisco'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
