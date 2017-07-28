package null

import (
	"github.com/cstuntz/packer/packer"
	"testing"
)

func TestNullArtifact(t *testing.T) {
	var _ packer.Artifact = new(NullArtifact)
}
