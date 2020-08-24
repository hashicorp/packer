package driver

import "testing"

func TestLibraryFilePath(t *testing.T) {
	tc := []struct {
		filePath        string
		libraryName     string
		libraryItemName string
		fileName        string
		valid           bool
	}{
		{
			filePath:        "lib/item/file",
			libraryName:     "lib",
			libraryItemName: "item",
			fileName:        "file",
			valid:           true,
		},
		{
			filePath:        "/lib/item/file",
			libraryName:     "lib",
			libraryItemName: "item",
			fileName:        "file",
			valid:           true,
		},
		{
			filePath: "/lib/item/filedir/file",
			valid:    false,
		},
		{
			filePath: "/lib/item",
			valid:    false,
		},
		{
			filePath: "/lib",
			valid:    false,
		},
	}

	for _, c := range tc {
		libraryFilePath := &LibraryFilePath{path: c.filePath}
		if err := libraryFilePath.Validate(); err != nil {
			if c.valid {
				t.Fatalf("Expecting %s to be valid", c.filePath)
			}
			continue
		}
		libraryName := libraryFilePath.GetLibraryName()
		if libraryName != c.libraryName {
			t.Fatalf("Expecting %s but got %s", c.libraryName, libraryName)
		}
		libraryItemName := libraryFilePath.GetLibraryItemName()
		if libraryItemName != c.libraryItemName {
			t.Fatalf("Expecting %s but got %s", c.libraryItemName, libraryItemName)
		}
		fileName := libraryFilePath.GetFileName()
		if fileName != c.fileName {
			t.Fatalf("Expecting %s but got %s", c.fileName, fileName)
		}
	}
}
