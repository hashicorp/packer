package plugingetter

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// A ChecksumError is returned when a checksum differs
type ChecksumError struct {
	Hash     hash.Hash
	Actual   []byte
	Expected []byte
	File     string
}

func (cerr *ChecksumError) Error() string {
	if cerr == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"Checksums did not match for %s.\nExpected: %s\nGot: %s\n%T",
		cerr.File,
		hex.EncodeToString(cerr.Expected),
		hex.EncodeToString(cerr.Actual),
		cerr.Hash, // ex: *sha256.digest
	)
}

type Checksummer struct {
	// Something like md5 or sha256
	Type string
	// Hash function
	hash.Hash
}

func (c *Checksummer) FileExt() string {
	return "_" + strings.ToUpper(c.Type) + "SUM"
}

// Checksum first reads the checksum in file `filePath + c.FileExt()`, then
// compares it to the checksum of the file in filePath.
func (c *Checksummer) Checksum(filePath string) error {
	checksumFile := filePath + c.FileExt()
	expected, err := ioutil.ReadFile(checksumFile)
	if err != nil {
		return fmt.Errorf("Checksum: failed to read checksum file: %s", err)
	}
	expected, err = hex.DecodeString(string(expected))
	if err != nil {
		return fmt.Errorf("Checksum(%q): invalid checksum: %s", checksumFile, err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Checksum: failed to open file for checksum: %s", err)
	}
	defer f.Close()

	c.Hash.Reset()
	if _, err := io.Copy(c.Hash, f); err != nil {
		return fmt.Errorf("Failed to hash: %s", err)
	}

	if actual := c.Hash.Sum(nil); !bytes.Equal(actual, expected) {
		return &ChecksumError{
			Hash:     c.Hash,
			Actual:   actual,
			Expected: expected,
			File:     filePath,
		}
	}

	return nil
}
