package common

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// CopyFile essentially replicates the `cp` command, for a file only.
//
// # Permissions are copied over from the source to destination
//
// The function detects if destination is a directory or a file (existent or not).
//
// If this is the former, we append the source file's basename to the
// directory and create the file from that inferred path.
func CopyFile(t *testing.T, dest, src string) {
	st, err := os.Stat(src)
	if err != nil {
		t.Fatalf("failed to stat origin file %q: %s", src, err)
	}

	// If the stat call fails, we assume dest is the destination file.
	dstStat, err := os.Stat(dest)
	if err == nil && dstStat.IsDir() {
		dest = filepath.Join(dest, filepath.Base(src))
	}

	destFD, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, st.Mode().Perm())
	if err != nil {
		t.Fatalf("failed to create cp destination file %q: %s", dest, err)
	}
	defer destFD.Close()

	srcFD, err := os.Open(src)
	if err != nil {
		t.Fatalf("failed to open source file to copy: %s", err)
	}
	defer srcFD.Close()

	_, err = io.Copy(destFD, srcFD)
	if err != nil {
		t.Fatalf("failed to copy from %q -> %q: %s", src, dest, err)
	}
}

// WriteFile writes `content` to a file `dest`
//
// The default permissions of that file is 0644
func WriteFile(t *testing.T, dest string, content string) {
	outFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("failed to open/create %q: %s", dest, err)
	}
	defer outFile.Close()

	_, err = fmt.Fprintf(outFile, content)
	if err != nil {
		t.Fatalf("failed to write to file %q: %s", dest, err)
	}
}

// TempWorkdir creates a working directory for a Packer test with the list of files
// given as input.
//
// The files should either have a path relative to the test that invokes it, or should
// be absolute.
// Each file will be copied to the root of the workdir being created.
//
// If any file cannot be found, this function will fail
func TempWorkdir(t *testing.T, files ...string) (string, func()) {
	var err error
	tempDir, err := os.MkdirTemp("", "packer-test-workdir-")
	if err != nil {
		t.Fatalf("failed to create temporary working directory: %s", err)
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tempDir)
			t.Errorf("failed to create temporary workdir: %s", err)
		}
	}()

	for _, file := range files {
		CopyFile(t, tempDir, file)
	}

	return tempDir, func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("failed to remove temporary workdir %q: %s. This will need manual action.", tempDir, err)
		}
	}
}

// SHA256Sum computes the SHA256 digest for an input file
//
// The digest is returned as a hexstring
func SHA256Sum(t *testing.T, file string) string {
	fl, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to compute sha256sum for %q: %s", file, err)
	}
	sha := sha256.New()
	sha.Write(fl)
	return fmt.Sprintf("%x", sha.Sum([]byte{}))
}

// currentDir returns the directory in which the current file is located.
//
// Since we're in tests it's reliable as they're supposed to run on the same
// machine the binary's compiled from, but goes to say it's not meant for use
// in distributed binaries.
func currentDir() (string, error) {
	// pc uintptr, file string, line int, ok bool
	_, testDir, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("couldn't get the location of the test suite file")
	}

	return filepath.Dir(testDir), nil
}
