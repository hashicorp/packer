package iso

import (
	"bytes"
	vmwcommon "github.com/cstuntz/packer/builder/vmware/common"
	"github.com/cstuntz/packer/packer"
	"github.com/mitchellh/multistep"
	"testing"
)

func testState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("driver", new(vmwcommon.DriverMock))
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}
