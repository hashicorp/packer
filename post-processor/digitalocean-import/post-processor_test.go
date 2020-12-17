package digitaloceanimport

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packersdk.PostProcessor = new(PostProcessor)
}

func TestPostProcessor_ImageArtifactExtraction(t *testing.T) {
	tt := []struct {
		Name          string
		Source        string
		Artifacts     []string
		ExpectedError string
	}{
		{Name: "EmptyArtifacts", ExpectedError: "no artifacts were provided"},
		{Name: "SingleArtifact", Source: "Sample.img", Artifacts: []string{"Sample.img"}},
		{Name: "SupportedArtifact", Source: "Example.tar.xz", Artifacts: []string{"Sample", "SomeZip.zip", "Example.tar.xz"}},
		{Name: "FirstSupportedArtifact", Source: "SomeVMDK.vmdk", Artifacts: []string{"Sample", "SomeVMDK.vmdk", "Example.xz"}},
		{Name: "NonSupportedArtifact", Artifacts: []string{"Sample", "SomeZip.zip", "Example.xz"}, ExpectedError: "no valid image file found"},
	}

	for _, tc := range tt {
		tc := tc
		source, err := extractImageArtifact(tc.Artifacts)

		if tc.Source != source {
			t.Errorf("expected the source to be %q, but got %q", tc.Source, source)
		}

		if err != nil && (tc.ExpectedError != err.Error()) {
			t.Errorf("unexpected error received; expected %q, but got %q", tc.ExpectedError, err.Error())
		}
	}
}
