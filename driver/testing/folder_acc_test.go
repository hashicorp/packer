package testing

import "testing"

func TestFolderAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	f, err := d.FindFolder("folder1/folder2")
	if err != nil {
		t.Fatalf("Cannot find the default folder '%v': %v", "folder1/folder2", err)
	}
	CheckFolderPath(t, f, "folder1/folder2")
}
