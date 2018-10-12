package hcloud

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packer.Artifact = (*Artifact)(nil)
}

func TestArtifactId(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, nil}
	expected := "42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, nil}
	expected := "A snapshot was created: 'packer-foobar' (ID: 42)"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
