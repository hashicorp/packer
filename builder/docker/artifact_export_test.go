package docker

import (
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestExportArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(ExportArtifact)
}
