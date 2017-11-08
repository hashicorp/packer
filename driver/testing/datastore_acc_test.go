package testing

import (
	"testing"
)

func TestDatastoreAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	ds, err := d.FindDatastore("datastore1")
	if err != nil {
		t.Fatalf("Cannot find the default datastore '%v': %v", "datastore1", err)
	}
	CheckDatastoreName(t, ds, "datastore1")
}
