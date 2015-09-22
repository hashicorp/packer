package brkt

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestBuilder_Impl(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should implement the packer.Builder interface")
	}
}
