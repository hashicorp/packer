package scaleway

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
	a := &Artifact{"packer-foobar-image", "cc586e45-5156-4f71-b223-cf406b10dd1d", "packer-foobar-snapshot", "cc586e45-5156-4f71-b223-cf406b10dd1c", "ams1", nil}
	expected := "ams1:cc586e45-5156-4f71-b223-cf406b10dd1d"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"packer-foobar-image", "cc586e45-5156-4f71-b223-cf406b10dd1d", "packer-foobar-snapshot", "cc586e45-5156-4f71-b223-cf406b10dd1c", "ams1", nil}
	expected := "An image was created: 'packer-foobar-image' (ID: cc586e45-5156-4f71-b223-cf406b10dd1d) in region 'ams1' based on snapshot 'packer-foobar-snapshot' (ID: cc586e45-5156-4f71-b223-cf406b10dd1c)"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
