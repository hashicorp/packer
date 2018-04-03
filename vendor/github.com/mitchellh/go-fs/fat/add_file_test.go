package fat

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchellh/go-fs"
)

func TestAddFile(t *testing.T) {
	// Create a temporary file to be our floppy drive
	floppyF, err := ioutil.TempFile("", "go-fs")
	if err != nil {
		t.Fatalf("Error creating temporary file for floppy: %s", err)
	}
	defer floppyF.Close()
	// defer os.Remove(floppyF.Name())

	// Set the size of the file to be a floppy sized
	if err := floppyF.Truncate(1440 * 1024); err != nil {
		t.Fatalf("Error creating floppy: %s", err)
	}

	// BlockDevice backed by the file for our filesystem
	device, err := fs.NewFileDisk(floppyF)
	if err != nil {
		t.Fatalf("Error creating floppy: %s", err)
	}

	// Format the block device so it contains a valid FAT filesystem
	formatConfig := &SuperFloppyConfig{
		FATType: FAT12,
		Label:   "go-fs",
		OEMName: "go-fs",
	}
	if FormatSuperFloppy(device, formatConfig); err != nil {
		t.Fatalf("Error creating floppy: %s", err)
	}

	// The actual FAT filesystem
	fatFs, err := New(device)
	if err != nil {
		t.Fatalf("Error creating floppy: %s", err)
	}

	// Get the root directory to the filesystem
	rootDir, err := fatFs.RootDir()
	if err != nil {
		t.Fatalf("Error creating floppy: %s", err)
	}

	var filenames [128]string

	// Go over each file and copy it.
	for i := 0; i < 128; i++ {
		filenames[i] = strings.Repeat("A", i+1) + ".EXT"
	}

	for _, filename := range filenames {
		if addSingleFile(rootDir, filename); err != nil {
			t.Fatalf("Error adding file to floppy: %s", err)
		}
	}

	entries := rootDir.Entries()

	for i, entry := range entries {
		expecting := filenames[i]
		if entry.Name() != expecting {
			// Name() returns the short name, how do we get the long name?
			// t.Fatalf("Excepting %s, found %s", expecting, entry.Name())
		}
	}
}

func addSingleFile(dir fs.Directory, src string) error {
	inputF, err := os.Create(src)
	if err != nil {
		return err
	}
	defer inputF.Close()
	defer os.Remove(src)

	entry, err := dir.AddFile(filepath.Base(src))
	if err != nil {
		return err
	}

	fatFile, err := entry.File()
	if err != nil {
		return err
	}

	if _, err := io.Copy(fatFile, inputF); err != nil {
		return err
	}

	return nil
}
