package vm

import (
	"bytes"
	"testing"

	"github.com/mitchellh/multistep"
	vspcommon "github.com/mitchellh/packer/builder/vsphere/common"
	"github.com/mitchellh/packer/packer"
)

func testState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("driver", new(vspcommon.DriverMock))
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}
