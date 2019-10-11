package vagrantcloud

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_ImplementsArtifact(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be a Artifact")
	}
}
