package chroot

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCopyFile(t *testing.T) {
	first, err := ioutil.TempFile("", "copy_files_test")
	if err != nil {
		t.Fatalf("couldn't create temp file.")
	}
	defer os.Remove(first.Name())
	newName := first.Name() + "-new"

	payload := "copy_files_test.go payload"
	if _, err = first.WriteString(payload); err != nil {
		t.Fatalf("Couldn't write payload to first file.")
	}

	if err := copySingle(newName, first.Name(), "cp"); err != nil {
		t.Fatalf("Couldn't copy file")
	}
	defer os.Remove(newName)

	second, err := os.Open(newName)
	if err != nil {
		t.Fatalf("Couldn't open copied file.")
	}
	defer second.Close()

	var copiedPayload = make([]byte, len(payload))
	if _, err := second.Read(copiedPayload); err != nil {
		t.Fatalf("Couldn't open copied file for reading.")
	}

	if string(copiedPayload) != payload {
		t.Fatalf("payload not copied.")
	}
}
