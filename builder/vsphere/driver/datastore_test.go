package driver

import "testing"

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
