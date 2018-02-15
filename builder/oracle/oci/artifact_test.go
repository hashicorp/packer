package oci

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifactImpl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}
