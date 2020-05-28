package getter

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	urlhelper "github.com/hashicorp/go-getter/v2/helper/url"
)

// FileChecksum helps verifying the checksum for a file.
type FileChecksum struct {
	Type     string
	Hash     hash.Hash
	Value    []byte
	Filename string
}

// String returns the hash type and the hash separated by a colon, for example:
//  "md5:090992ba9fd140077b0661cb75f7ce13"
//  "sha1:ebfb681885ddf1234c18094a45bbeafd91467911"
func (c *FileChecksum) String() string {
	return c.Type + ":" + hex.EncodeToString(c.Value)
}

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

// Checksum computes the Checksum for filePath using the hashing algorithm from
// c.Hash and compares it to c.Value. If those values differ a ChecksumError
// will be returned.
func (c *FileChecksum) Checksum(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed to open file for checksum: %s", err)
	}
	defer f.Close()

	c.Hash.Reset()
	if _, err := io.Copy(c.Hash, f); err != nil {
		return fmt.Errorf("Failed to hash: %s", err)
	}

	if actual := c.Hash.Sum(nil); !bytes.Equal(actual, c.Value) {
		return &ChecksumError{
			Hash:     c.Hash,
			Actual:   actual,
			Expected: c.Value,
			File:     filePath,
		}
	}

	return nil
}

// GetChecksum extracts the checksum from the `checksum` parameter
// of the src of the Request
// ex:
//  http://hashicorp.com/terraform?checksum=<checksumValue>
//  http://hashicorp.com/terraform?checksum=<checksumType>:<checksumValue>
//  http://hashicorp.com/terraform?checksum=file:<checksum_url>
// when the checksum is in a file, GetChecksum will first client.Get it
// in a temporary directory, parse the content of the file and finally delete it.
// The content of a checksum file is expected to be BSD style or GNU style.
// For security reasons GetChecksum does not try to get the current working directory
// and as a result, relative files will only be found when Request.Pwd is set.
//
// BSD-style checksum:
//  MD5 (file1) = <checksum>
//  MD5 (file2) = <checksum>
//
// GNU-style:
//  <checksum>  file1
//  <checksum> *file2
func (c *Client) GetChecksum(ctx context.Context, req *Request) (*FileChecksum, error) {
	var err error
	if req.u == nil {
		req.u, err = urlhelper.Parse(req.Src)
		if err != nil {
			return nil, err
		}
	}
	q := req.u.Query()
	v := q.Get("checksum")

	if v == "" {
		return nil, nil
	}

	vs := strings.SplitN(v, ":", 2)
	switch len(vs) {
	case 2:
		break // good
	default:
		// here, we try to guess the checksum from it's length
		// if the type was not passed
		return newChecksumFromValue(v, filepath.Base(req.u.EscapedPath()))
	}

	checksumType, checksumValue := vs[0], vs[1]

	switch checksumType {
	case "file":
		return c.checksumFromFile(ctx, checksumValue, req.u.Path, req.Pwd)
	default:
		return newChecksumFromType(checksumType, checksumValue, filepath.Base(req.u.EscapedPath()))
	}
}

func newChecksum(checksumValue, filename string) (*FileChecksum, error) {
	c := &FileChecksum{
		Filename: filename,
	}
	var err error
	c.Value, err = hex.DecodeString(checksumValue)
	if err != nil {
		return nil, fmt.Errorf("invalid checksum: %s", err)
	}
	return c, nil
}

func newChecksumFromType(checksumType, checksumValue, filename string) (*FileChecksum, error) {
	c, err := newChecksum(checksumValue, filename)
	if err != nil {
		return nil, err
	}

	c.Type = strings.ToLower(checksumType)
	switch c.Type {
	case "md5":
		c.Hash = md5.New()
	case "sha1":
		c.Hash = sha1.New()
	case "sha256":
		c.Hash = sha256.New()
	case "sha512":
		c.Hash = sha512.New()
	default:
		return nil, fmt.Errorf(
			"unsupported checksum type: %s", checksumType)
	}

	return c, nil
}

func newChecksumFromValue(checksumValue, filename string) (*FileChecksum, error) {
	c, err := newChecksum(checksumValue, filename)
	if err != nil {
		return nil, err
	}

	switch len(c.Value) {
	case md5.Size:
		c.Hash = md5.New()
		c.Type = "md5"
	case sha1.Size:
		c.Hash = sha1.New()
		c.Type = "sha1"
	case sha256.Size:
		c.Hash = sha256.New()
		c.Type = "sha256"
	case sha512.Size:
		c.Hash = sha512.New()
		c.Type = "sha512"
	default:
		return nil, fmt.Errorf("Unknown type for checksum %s", checksumValue)
	}

	return c, nil
}

// checksumFromFile will return the first file checksum found in the
// `checksumURL` file that corresponds to the `checksummedPath` path.
//
// checksumFromFile will infer the hashing algorithm based on the checksumURL
// file content.
//
// checksumFromFile will only return checksums for files that match
// checksummedPath, which is the object being checksummed.
func (c *Client) checksumFromFile(ctx context.Context, checksumURL string, checksummedPath string, pwd string) (*FileChecksum, error) {
	checksumFileURL, err := urlhelper.Parse(checksumURL)
	if err != nil {
		return nil, err
	}

	tempfile, err := tmpFile("", filepath.Base(checksumFileURL.Path))
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempfile)

	req := &Request{
		Pwd:  pwd,
		Mode: ModeFile,
		Src:  checksumURL,
		Dst:  tempfile,
		// ProgressListener: c.ProgressListener, TODO(adrien): pass progress bar ?
	}

	if _, err = c.Get(ctx, req); err != nil {
		return nil, fmt.Errorf(
			"Error downloading checksum file: %s", err)
	}

	filename := filepath.Base(checksummedPath)
	absPath, err := filepath.Abs(checksummedPath)
	if err != nil {
		return nil, err
	}
	checksumFileDir := filepath.Dir(checksumFileURL.Path)
	relpath, err := filepath.Rel(checksumFileDir, absPath)
	switch {
	case err == nil ||
		err.Error() == "Rel: can't make "+absPath+" relative to "+checksumFileDir:
		// ex: on windows C:\gopath\...\content.txt cannot be relative to \
		// which is okay, may be another expected path will work.
		break
	default:
		return nil, err
	}

	// possible file identifiers:
	options := []string{
		filename,       // ubuntu-14.04.1-server-amd64.iso
		"*" + filename, // *ubuntu-14.04.1-server-amd64.iso  Standard checksum
		"?" + filename, // ?ubuntu-14.04.1-server-amd64.iso  shasum -p
		relpath,        // dir/ubuntu-14.04.1-server-amd64.iso
		"./" + relpath, // ./dir/ubuntu-14.04.1-server-amd64.iso
		absPath,        // fullpath; set if local
	}

	f, err := os.Open(tempfile)
	if err != nil {
		return nil, fmt.Errorf(
			"Error opening downloaded file: %s", err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf(
					"Error reading checksum file: %s", err)
			}
			break
		}
		checksum, err := parseChecksumLine(line)
		if err != nil || checksum == nil {
			continue
		}
		if checksum.Filename == "" {
			// filename not sure, let's try
			return checksum, nil
		}
		// make sure the checksum is for the right file
		for _, option := range options {
			if option != "" && checksum.Filename == option {
				// any checksum will work so we return the first one
				return checksum, nil
			}
		}
		// The checksum filename can contain a sub folder to differ versions.
		// e.g. ./netboot/mini.iso and ./hwe-netboot/mini.iso
		// In this case we remove root folder characters to compare with the checksummed path
		fn := strings.TrimLeft(checksum.Filename, "./")
		if strings.Contains(checksummedPath, fn) {
			return checksum, nil
		}
	}
	return nil, fmt.Errorf("no checksum found in: %s", checksumURL)
}

// parseChecksumLine takes a line from a checksum file and returns
// checksumType, checksumValue and filename parseChecksumLine guesses the style
// of the checksum BSD vs GNU by splitting the line and by counting the parts.
// of a line.
// for BSD type sums parseChecksumLine guesses the hashing algorithm
// by checking the length of the checksum.
func parseChecksumLine(line string) (*FileChecksum, error) {
	parts := strings.Fields(line)

	switch len(parts) {
	case 4:
		// BSD-style checksum:
		//  MD5 (file1) = <checksum>
		//  MD5 (file2) = <checksum>
		if len(parts[1]) <= 2 ||
			parts[1][0] != '(' || parts[1][len(parts[1])-1] != ')' {
			return nil, fmt.Errorf(
				"Unexpected BSD-style-checksum filename format: %s", line)
		}
		filename := parts[1][1 : len(parts[1])-1]
		return newChecksumFromType(parts[0], parts[3], filename)
	case 2:
		// GNU-style:
		//  <checksum>  file1
		//  <checksum> *file2
		return newChecksumFromValue(parts[0], parts[1])
	case 0:
		return nil, nil // empty line
	default:
		return newChecksumFromValue(parts[0], "")
	}
}
