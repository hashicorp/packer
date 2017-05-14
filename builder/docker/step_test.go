package docker

import (
	"bytes"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"testing"
)

func testState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("config", testConfigStruct(t))
	state.Put("driver", &MockDriver{})
	state.Put("hook", &packer.MockHook{})
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}
