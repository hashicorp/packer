package digitaloceanimport

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}

func TestPostProcsor_extractImageArtifact(t *testing.T) {
	tt := []struct {
		Name      string
		Source    string
		Artifacts []string
	}{
		{Name: "EmptyArtifacts"},
		{Name: "SingleArtifact", Source: "Sample.img", Artifacts: []string{"Sample.img"}},
		{Name: "SupportedArtifact", Source: "Example.tar.xz", Artifacts: []string{"Sample", "SomeZip.zip", "Example.tar.xz"}},
		{Name: "FirstSupportedArtifact", Source: "SomeVMDK.vmdk", Artifacts: []string{"Sample", "SomeVMDK.vmdk", "Example.xz"}},
		{Name: "NonSupportedArtifact", Artifacts: []string{"Sample", "SomeZip.zip", "Example.xz"}},
	}

	for _, tc := range tt {
		tc := tc
		source, _ := extractImageArtifact(tc.Artifacts)

		if tc.Source != source {
			t.Errorf("expected the source to be %q, but got %q", tc.Source, source)
		}
	}
}
