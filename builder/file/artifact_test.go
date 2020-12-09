package file

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestNullArtifact(t *testing.T) {
	var _ packersdk.Artifact = new(FileArtifact)
}
