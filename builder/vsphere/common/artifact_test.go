package common

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatal("Artifact must be a proper artifact")
	}
}
