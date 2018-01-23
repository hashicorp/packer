package driver

import (
	"testing"
)

func TestHostAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := newTestDriver(t)
	host, err := d.FindHost(hostName)
	if err != nil {
		t.Fatalf("Cannot find the default host '%v': %v", "datastore1", err)
	}

	info, err := host.Info("name")
	if err != nil {
		t.Fatalf("Cannot read host properties: %v", err)
	}
	if info.Name != hostName {
		t.Errorf("Wrong host name: expected '%v', got: '%v'", hostName, info.Name)
	}
}
