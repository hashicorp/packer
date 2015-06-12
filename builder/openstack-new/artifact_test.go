package openstack

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}

func TestArtifactId(t *testing.T) {
	expected := `b8cdf55b-c916-40bd-b190-389ec144c4ed`

	a := &Artifact{
		ImageId: "b8cdf55b-c916-40bd-b190-389ec144c4ed",
	}

	result := a.Id()
	if result != expected {
		t.Fatalf("bad: %s", result)
	}
}

func TestArtifactString(t *testing.T) {
	expected := "An image was created: b8cdf55b-c916-40bd-b190-389ec144c4ed"

	a := &Artifact{
		ImageId: "b8cdf55b-c916-40bd-b190-389ec144c4ed",
	}
	result := a.String()
	if result != expected {
		t.Fatalf("bad: %s", result)
	}
}
