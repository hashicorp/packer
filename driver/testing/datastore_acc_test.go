package testing

import (
	"testing"
)

var infoParams = [][]string {
	//{}, // FIXME: Doesn't work
	//{"*"}, // FIXME: Doesn't work
	{"host", "vm"}, // TODO: choose something more meaningful
}

func TestDatastore(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	ds, err := d.FindDatastore(DefaultDatastore)
	if err != nil {
		t.Fatalf("`FindDatatore` can't find default datastore '%v'. Error: %v",
			DefaultDatastore, err)
	}
	for _, params := range infoParams {
		_, err := ds.Info(params...)
		if err != nil {
			t.Errorf("Cannot read datastore properties with parameters %v: %v",
				params, err)
		}
	}
	// TODO: add more checks
}
