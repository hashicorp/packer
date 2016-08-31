package profitbricks

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

func TestArtifactString(t *testing.T) {
	a := &Artifact{"packer-foobar"}
	expected := "A snapshot was created: 'packer-foobar'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
