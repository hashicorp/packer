package vsphere

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
	artifact := NewArtifact("datastore", "vmfolder", "vmname", nil)
	if artifact.Id() != "datastore::vmfolder::vmname" {
		t.Fatalf("must return datastore, vmfolder and vmname splitted by :: as Id")
	}
}
