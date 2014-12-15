package packer

import (
	"testing"
)

func TestNullArtifact(t *testing.T) {
	var _ Artifact = new(NullArtifact)
}
