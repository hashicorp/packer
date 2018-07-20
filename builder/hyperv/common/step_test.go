package common

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func testState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("driver", new(DriverMock))
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}

// Generates an absolute path to a directory under OS temp with a name
// beginning with prefix and a UUID appended to the end
func genTestDirPath(prefix string) string {
	return filepath.Join(os.TempDir(), prefix+"-"+uuid.TimeOrderedUUID())
}
