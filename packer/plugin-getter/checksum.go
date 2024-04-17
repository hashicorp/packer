// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugingetter

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
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
		"Checksums (%T) did not match.\nExpected: %s\nGot     : %s\n",
		cerr.Hash, // ex: *sha256.digest
		hex.EncodeToString(cerr.Expected),
		hex.EncodeToString(cerr.Actual),
	)
}

type Checksum []byte

func (c Checksum) String() string { return hex.EncodeToString(c) }

type FileChecksum struct {
	Filename string
	Expected Checksum
	Checksummer
}

type Checksummer struct {
	// Something like md5 or sha256
	Type string
}

func (c *Checksummer) Hash() hash.Hash {
	switch c.Type {
	case "sha256":
		return sha256.New()
	case "md5":
		return md5.New()
	}
	panic(fmt.Sprintf("Unsupported hash type %q, only md5 and sha256 are supported", c.Type))
}

func (c *Checksummer) FileExt() string {
	return "_" + strings.ToUpper(c.Type) + "SUM"
}

// GetCacheChecksumOfFile will extract the checksum from file `filePath + c.FileExt()`.
// It expects the checksum file to only contains the checksum and nothing else.
func (c *Checksummer) GetCacheChecksumOfFile(filePath string) ([]byte, error) {
	checksumFile := filePath + c.FileExt()

	f, err := os.Open(checksumFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return c.ParseChecksum(f)
}

// ParseChecksum expects the checksum reader to only contain the checksum and
// nothing else.
func (c *Checksummer) ParseChecksum(f io.Reader) (Checksum, error) {
	hash := c.Hash()
	res := make([]byte, hash.Size())
	_, err := hex.NewDecoder(f).Read(res)
	if err == io.EOF {
		err = nil
	}
	return res, err
}

// ChecksumFile compares the expected checksum to the checksum of the file in
// filePath using the hash function.
func (c *Checksummer) ChecksumFile(expected []byte, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Checksum: failed to open file for checksum: %s", err)
	}
	defer f.Close()
	err = c.Checksum(expected, f)
	if cerr, ok := err.(*ChecksumError); ok {
		cerr.File = filePath
	}
	return err
}

func (c *Checksummer) Sum(f io.Reader) ([]byte, error) {
	hash := c.Hash()
	hash.Reset()
	if _, err := io.Copy(hash, f); err != nil {
		return nil, fmt.Errorf("Failed to hash: %s", err)
	}
	return hash.Sum(nil), nil
}

func (c *Checksummer) Checksum(expected []byte, f io.Reader) error {
	actual, err := c.Sum(f)
	if err != nil {
		return err
	}

	if !bytes.Equal(actual, expected) {
		return &ChecksumError{
			Hash:     c.Hash(),
			Actual:   actual,
			Expected: expected,
		}
	}

	return nil
}
