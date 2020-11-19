package none

import (
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestCommIsCommunicator(t *testing.T) {
	var raw interface{}
	raw = &comm{}
	if _, ok := raw.(packersdk.Communicator); !ok {
		t.Fatalf("comm must be a communicator")
	}
}
