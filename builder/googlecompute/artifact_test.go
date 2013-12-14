package googlecompute

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}
