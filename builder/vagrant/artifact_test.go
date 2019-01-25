package vagrant

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{} = &artifact{}

	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact does not implement packer.Artifact")
	}
}

func TestArtifactId(t *testing.T) {
	a := &artifact{
		OutputDir: "/my/dir",
		BoxName:   "package.box",
	}

	expected := "/my/dir/package.box"
	if strings.Compare(a.Id(), expected) != 0 {
		t.Fatalf("artifact ID should match: expected: %s received: %s", expected, a.Id())
	}
}

func TestArtifactString(t *testing.T) {
	a := &artifact{
		OutputDir: "/my/dir",
		BoxName:   "package.box",
	}
	expected := "Vagrant box is /my/dir/package.box"

	if strings.Compare(a.String(), expected) != 0 {
		t.Fatalf("artifact string should match: expected: %s received: %s", expected, a.String())
	}
}
