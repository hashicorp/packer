package docker

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestExportArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(ExportArtifact)
}
