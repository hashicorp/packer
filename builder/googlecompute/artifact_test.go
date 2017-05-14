package googlecompute

import (
	"github.com/hashicorp/packer/packer"
	"testing"
)

func TestArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}
