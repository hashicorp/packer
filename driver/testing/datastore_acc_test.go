package testing

import (
	"testing"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

func TestDatastoreAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	ds, err := d.FindDatastore(DefaultDatastore)
	if err != nil {
		t.Fatalf("Cannot find default datastore '%v': %v", DefaultDatastore, err)
	}
	CheckDatastoreName(t, ds, DefaultDatastore)
}
