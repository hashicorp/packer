package chroot

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func TestCopyFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

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
	first.Sync()

	cmd := ShellCommand(fmt.Sprintf("cp %s %s", first.Name(), newName))
	if err := cmd.Run(); err != nil {
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
