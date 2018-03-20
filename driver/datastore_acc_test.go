package driver

import (
	"testing"
)

func TestDatastoreAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := newTestDriver(t)
	ds, err := d.FindDatastore("datastore1", "")
	if err != nil {
		t.Fatalf("Cannot find the default datastore '%v': %v", "datastore1", err)
	}
	info, err := ds.Info("name")
	if err != nil {
		t.Fatalf("Cannot read datastore properties: %v", err)
	}
	if info.Name != "datastore1" {
		t.Errorf("Wrong datastore. expected: 'datastore1', got: '%v'", info.Name)
	}
}
