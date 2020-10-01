package driver

import (
	"testing"

	"github.com/vmware/govmomi/simulator"
)

func TestDatastoreIsoPath(t *testing.T) {
	tc := []struct {
		isoPath  string
		filePath string
		valid    bool
	}{
		{
			isoPath:  "[datastore] dir/subdir/file",
			filePath: "dir/subdir/file",
			valid:    true,
		},
		{
			isoPath:  "[] dir/subdir/file",
			filePath: "dir/subdir/file",
			valid:    true,
		},
		{
			isoPath:  "dir/subdir/file",
			filePath: "dir/subdir/file",
			valid:    true,
		},
		{
			isoPath:  "[datastore] /dir/subdir/file",
			filePath: "/dir/subdir/file",
			valid:    true,
		},
		{
			isoPath: "/dir/subdir/file [datastore] ",
			valid:   false,
		},
		{
			isoPath: "[datastore][] /dir/subdir/file",
			valid:   false,
		},
		{
			isoPath: "[data/store] /dir/subdir/file",
			valid:   false,
		},
		{
			isoPath:  "[data store] /dir/sub dir/file",
			filePath: "/dir/sub dir/file",
			valid:    true,
		},
		{
			isoPath:  "   [datastore] /dir/subdir/file",
			filePath: "/dir/subdir/file",
			valid:    true,
		},
		{
			isoPath:  "[datastore]    /dir/subdir/file",
			filePath: "/dir/subdir/file",
			valid:    true,
		},
		{
			isoPath:  "[datastore] /dir/subdir/file     ",
			filePath: "/dir/subdir/file",
			valid:    true,
		},
		{
			isoPath:  "[привѣ́тъ] /привѣ́тъ/привѣ́тъ/привѣ́тъ",
			filePath: "/привѣ́тъ/привѣ́тъ/привѣ́тъ",
			valid:    true,
		},
		// Test case for #9846
		{
			isoPath:  "[ISO-StorageLun9] Linux/rhel-8.0-x86_64-dvd.iso",
			filePath: "Linux/rhel-8.0-x86_64-dvd.iso",
			valid:    true,
		},
	}

	for i, c := range tc {
		dsIsoPath := &DatastoreIsoPath{path: c.isoPath}
		if dsIsoPath.Validate() != c.valid {
			t.Fatalf("%d Expecting %s to be %t but was %t", i, c.isoPath, c.valid, !c.valid)
		}
		if !c.valid {
			continue
		}
		filePath := dsIsoPath.GetFilePath()
		if filePath != c.filePath {
			t.Fatalf("%d Expecting %s but got %s", i, c.filePath, filePath)
		}
	}
}

func TestVCenterDriver_FindDatastore(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	_, datastore := sim.ChooseSimulatorPreCreatedDatastore()
	_, host := sim.ChooseSimulatorPreCreatedHost()

	tc := []struct {
		name       string
		datastore  string
		host       string
		fail       bool
		errMessage string
	}{
		{
			name:      "should find datastore when name is provided",
			datastore: datastore.Name,
			fail:      false,
		},
		{
			name: "should find datastore when only host is provided",
			host: host.Name,
			fail: false,
		},
		{
			name:      "should not find invalid datastore",
			datastore: "invalid",
			fail:      true,
		},
		{
			name: "should not find invalid host",
			host: "invalid",
			fail: true,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			ds, err := sim.driver.FindDatastore(c.datastore, c.host)
			if c.fail {
				if err == nil {
					t.Fatalf("expected to fail")
				}
				if c.errMessage != "" && err.Error() != c.errMessage {
					t.Fatalf("unexpected error message %s", err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("should not fail: %s", err.Error())
				}
				if ds == nil {
					t.Fatalf("expected to find datastore")
				}
			}
		})
	}
}

func TestVCenterDriver_MultipleDatastoreError(t *testing.T) {
	model := simulator.ESX()
	model.Datastore = 2
	sim, err := NewCustomVCenterSimulator(model)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	_, host := sim.ChooseSimulatorPreCreatedHost()

	_, err = sim.driver.FindDatastore("", host.Name)
	if err == nil {
		t.Fatalf("expected to fail")
	}
	if err.Error() != "Host has multiple datastores. Specify it explicitly" {
		t.Fatalf("unexpected error message %s", err.Error())
	}
}
