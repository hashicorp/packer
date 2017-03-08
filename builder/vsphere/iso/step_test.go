package iso

import (
	"bytes"
	vspcommon "github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"testing"
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
