package vagrant

import (
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestArtifact_ImplementsArtifact(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packersdk.Artifact); !ok {
		t.Fatalf("Artifact should be a Artifact")
	}
}

func TestArtifact_Id(t *testing.T) {
	artifact := NewArtifact("vmware", "./")
	if artifact.Id() != "vmware" {
		t.Fatalf("should return name as Id")
	}
}
