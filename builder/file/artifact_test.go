package file

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestNullArtifact(t *testing.T) {
	var _ packer.Artifact = new(FileArtifact)
}
