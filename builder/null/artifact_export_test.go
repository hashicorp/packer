package null

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestNullArtifact(t *testing.T) {
	var _ packer.Artifact = new(NullArtifact)
}
