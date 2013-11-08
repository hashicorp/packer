package vmware

import (
	"testing"
)

func TestESX5Driver_implDriver(t *testing.T) {
	var _ Driver = new(ESX5Driver)
}

func TestESX5Driver_implRemoteDriver(t *testing.T) {
	var _ RemoteDriver = new(ESX5Driver)
}
