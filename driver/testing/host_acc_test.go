package testing

import (
	"testing"
)

func TestHostAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	host, err := d.FindHost(TestHost)
	if err != nil {
		t.Fatalf("Cannot find the default datastore '%v': %v", TestDatastore, err)
	}
	switch info, err := host.Info("name"); {
	case err != nil:
		t.Errorf("Cannot read datastore properties: %v", err)
	case info.Name != TestHost:
		t.Errorf("Wrong datastore. expected: %v, got: %v", TestHost, info.Name)
	}
}
