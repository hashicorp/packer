package validate

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestCommand_Impl(t *testing.T) {
	var raw interface{}
	raw = new(Command)
	if _, ok := raw.(packer.Command); !ok {
		t.Fatalf("must be a Command")
	}
}
