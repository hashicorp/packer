package common

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &RemoteArtifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatal("Artifact must be a proper artifact")
	}
}
