package file

import (
	"testing"

	"github.com/cstuntz/packer/packer"
)

func TestNullArtifact(t *testing.T) {
	var _ packer.Artifact = new(FileArtifact)
}
