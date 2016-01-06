package vagrant

import (
	"strings"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestDigitalOceanProvider_impl(t *testing.T) {
	var _ Provider = new(DigitalOceanProvider)
}

func TestDigitalOceanProvider_KeepInputArtifact(t *testing.T) {
	p := new(DigitalOceanProvider)

	if !p.KeepInputArtifact() {
		t.Fatal("should keep input artifact")
	}
}

func TestDigitalOceanProvider_ArtifactId(t *testing.T) {
	p := new(DigitalOceanProvider)
	ui := testUi()
	artifact := &packer.MockArtifact{
		IdValue: "San Francisco:42",
	}

	vagrantfile, _, err := p.Process(ui, artifact, "foo")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	image := `digital_ocean.image = "42"`
	if strings.Index(vagrantfile, image) == -1 {
		t.Fatalf("wrong image substitution: %s", vagrantfile)
	}
	region := `digital_ocean.region = "San Francisco"`
	if strings.Index(vagrantfile, region) == -1 {
		t.Fatalf("wrong region substitution: %s", vagrantfile)
	}
}
