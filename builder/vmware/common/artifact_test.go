package common

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestLocalArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(artifact)
}
