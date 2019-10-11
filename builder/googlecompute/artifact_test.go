package googlecompute

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}
