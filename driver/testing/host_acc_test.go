package testing

import (
	"testing"
)

func TestHostAcc(t *testing.T) {
	initDriverAcceptanceTest(t)
	hostName := "esxi-1.vsphere55.test"

	d := NewTestDriver(t)
	host, err := d.FindHost(hostName)
	if err != nil {
		t.Fatalf("Cannot find the default host '%v': %v", "datastore1", err)
	}
	switch info, err := host.Info("name"); {
	case err != nil:
		t.Errorf("Cannot read host properties: %v", err)
	case info.Name != hostName:
		t.Errorf("Wrong host name: expected '%v', got: '%v'", hostName, info.Name)
	}
}
