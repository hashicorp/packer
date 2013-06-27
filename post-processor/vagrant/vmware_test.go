package vagrant

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestVMwareBoxPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var raw interface{}
	raw = &VMwareBoxPostProcessor{}
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatalf("VMware PostProcessor should be a PostProcessor")
	}
}
