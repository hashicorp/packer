package driver

import (
	"testing"
)

func TestHostAcc(t *testing.T) {
	d := newTestDriver(t)
	host, err := d.FindHost(TestHostName)
	if err != nil {
		t.Fatalf("Cannot find the default host '%v': %v", "datastore1", err)
	}

	info, err := host.Info("name")
	if err != nil {
		t.Fatalf("Cannot read host properties: %v", err)
	}
	if info.Name != TestHostName {
		t.Errorf("Wrong host name: expected '%v', got: '%v'", TestHostName, info.Name)
	}
}
