package lxc

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestCommunicator_ImplementsCommunicator(t *testing.T) {
	var raw interface{}
	raw = &LxcAttachCommunicator{}
	if _, ok := raw.(packer.Communicator); !ok {
		t.Fatalf("Communicator should be a communicator")
	}
}
