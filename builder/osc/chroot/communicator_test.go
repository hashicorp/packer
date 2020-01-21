package chroot

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestCommunicator_ImplementsCommunicator(t *testing.T) {
	var raw interface{}
	raw = &Communicator{}
	if _, ok := raw.(packer.Communicator); !ok {
		t.Fatalf("Communicator should be a communicator")
	}
}
