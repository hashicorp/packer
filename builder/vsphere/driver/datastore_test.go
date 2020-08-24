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
	}

	for _, c := range tc {
		dsIsoPath := &DatastoreIsoPath{path: c.isoPath}
		if dsIsoPath.Validate() != c.valid {
			t.Fatalf("Expecting %s to be %t but was %t", c.isoPath, c.valid, !c.valid)
		}
		if !c.valid {
			continue
		}
		filePath := dsIsoPath.GetFilePath()
		if filePath != c.filePath {
			t.Fatalf("Expecting %s but got %s", c.filePath, filePath)
		}
	}
}
