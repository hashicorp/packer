package iso

import (
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"testing"
)

func TestESX5Driver_implDriver(t *testing.T) {
	var _ vmwcommon.Driver = new(ESX5Driver)
}

func TestESX5Driver_implOutputDir(t *testing.T) {
	var _ vmwcommon.OutputDir = new(ESX5Driver)
}

func TestESX5Driver_implRemoteDriver(t *testing.T) {
	var _ RemoteDriver = new(ESX5Driver)
}
