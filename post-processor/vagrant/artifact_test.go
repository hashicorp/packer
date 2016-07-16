package vagrant

import (
	"github.com/mitchellh/packer/packer"
	"reflect"
	"testing"
)

func TestArtifact_ImplementsArtifact(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be a Artifact")
	}
}

func TestArtifact_Id(t *testing.T) {
	artifact := NewArtifact("vmware", "./")
	if artifact.Id() != "vmware" {
		t.Fatalf("should return name as Id")
	}
}

func TestArtifact_StateAtlasMetadata(t *testing.T) {
	artifact := NewArtifact("vmware", "./")

	actual := artifact.State("atlas.artifact.metadata")
	expected := map[string]string{"provider": "vmware_desktop"}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}
