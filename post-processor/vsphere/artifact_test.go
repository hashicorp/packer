package vsphere

import (
	"github.com/hashicorp/packer/packer"
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
	artifact := NewArtifact("datastore", "vmfolder", "vmname", nil)
	if artifact.Id() != "datastore::vmfolder::vmname" {
		t.Fatalf("must return datastore, vmfolder and vmname splitted by :: as Id")
	}
}
