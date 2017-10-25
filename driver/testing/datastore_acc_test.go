package testing

import (
	"testing"
)

func TestDatastoreAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	ds, err := d.FindDatastore(TestDatastore)
	if err != nil {
		t.Fatalf("Cannot find the default datastore '%v': %v", TestDatastore, err)
	}
	CheckDatastoreName(t, ds, TestDatastore)
}
