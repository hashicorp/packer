package vagrant

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestVBoxBoxPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var raw interface{}
	raw = &VBoxBoxPostProcessor{}
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatalf("VBox PostProcessor should be a PostProcessor")
	}
}
