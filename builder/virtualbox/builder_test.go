package virtualbox

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Error("Builder must implement builder.")
	}
}
