package bsusurrogate

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatal("Builder should be a builder")
	}
}
