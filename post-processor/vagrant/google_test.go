package vagrant

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestGoogleProvider_impl(t *testing.T) {
	var _ Provider = new(GoogleProvider)
}

func TestGoogleProvider_KeepInputArtifact(t *testing.T) {
	p := new(GoogleProvider)

	if !p.KeepInputArtifact() {
		t.Fatal("should keep input artifact")
	}
}

func TestGoogleProvider_ArtifactId(t *testing.T) {
	p := new(GoogleProvider)
	ui := testUi()
	artifact := &packer.MockArtifact{
		IdValue: "packer-1234",
	}

	vagrantfile, _, err := p.Process(ui, artifact, "foo")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	result := `google.image = "packer-1234"`
	if !strings.Contains(vagrantfile, result) {
		t.Fatalf("wrong substitution: %s", vagrantfile)
	}
}
