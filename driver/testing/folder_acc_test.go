package testing

import "testing"

func TestFolderAcc(t *testing.T) {
	initDriverAcceptanceTest(t)

	d := NewTestDriver(t)
	f, err := d.FindFolder(TestFolder)
	if err != nil {
		t.Fatalf("Cannot find the default folder '%v': %v", TestFolder, err)
	}
	CheckFolderPath(t, f, TestFolder)
}
