package common

import (
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestLocalArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(artifact)
}
