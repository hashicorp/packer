package docker

import (
	"github.com/cstuntz/packer/packer"
	"testing"
)

func TestExportArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(ExportArtifact)
}
