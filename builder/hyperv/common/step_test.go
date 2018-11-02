package common

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/configfile"
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

// Generates an absolute path to a directory with a name
// beginning with prefix and a UUID appended to the end
func genTestDirPath(prefix string) string {
	var suffix string

	if prefix == "" {
		suffix = uuid.TimeOrderedUUID()
	} else {
		suffix = prefix + "-" + uuid.TimeOrderedUUID()
	}

	tdprefix, _ := configfile.ConfigTmpDir()
	td, err := ioutil.TempDir(tdprefix, "hyperv")
	if err != nil {
		// use CWD as last-ditch
		td, err = filepath.Abs(".")
	}

	return filepath.Join(td, suffix)
}
