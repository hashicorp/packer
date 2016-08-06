package none

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestCommIsCommunicator(t *testing.T) {
	var raw interface{}
	raw = &comm{}
	if _, ok := raw.(packer.Communicator); !ok {
		t.Fatalf("comm must be a communicator")
	}
}

