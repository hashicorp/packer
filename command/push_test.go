package command

import (
	"fmt"
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestPush_noArgs(t *testing.T) {
	c := &PushCommand{Meta: testMeta(t)}
	code := c.Run(nil)
	if code != 1 {
		t.Fatalf("bad: %#v", code)
	}
}

func TestPush_multiArgs(t *testing.T) {
	c := &PushCommand{Meta: testMeta(t)}
	code := c.Run([]string{"one", "two"})
	if code != 1 {
		t.Fatalf("bad: %#v", code)
	}
}

func TestPush(t *testing.T) {
	var actualR io.Reader
	var actualOpts *uploadOpts
	uploadFn := func(r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		actualR = r
		actualOpts = opts

		doneCh := make(chan struct{})
		close(doneCh)
		return doneCh, nil, nil
	}

	c := &PushCommand{
		Meta:     testMeta(t),
		uploadFn: uploadFn,
	}

	args := []string{filepath.Join(testFixture("push"), "template.json")}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	actual := testArchive(t, actualR)
	expected := []string{
		archiveTemplateEntry,
		"template.json",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestPush_noName(t *testing.T) {
	uploadFn := func(r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		return nil, nil, nil
	}

	c := &PushCommand{
		Meta:     testMeta(t),
		uploadFn: uploadFn,
	}

	args := []string{filepath.Join(testFixture("push-no-name"), "template.json")}
	if code := c.Run(args); code != 1 {
		fatalCommand(t, c.Meta)
	}
}

func TestPush_uploadError(t *testing.T) {
	uploadFn := func(r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		return nil, nil, fmt.Errorf("bad")
	}

	c := &PushCommand{
		Meta:     testMeta(t),
		uploadFn: uploadFn,
	}

	args := []string{filepath.Join(testFixture("push"), "template.json")}
	if code := c.Run(args); code != 1 {
		fatalCommand(t, c.Meta)
	}
}

func TestPush_uploadErrorCh(t *testing.T) {
	uploadFn := func(r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("bad")
		return nil, errCh, nil
	}

	c := &PushCommand{
		Meta:     testMeta(t),
		uploadFn: uploadFn,
	}

	args := []string{filepath.Join(testFixture("push"), "template.json")}
	if code := c.Run(args); code != 1 {
		fatalCommand(t, c.Meta)
	}
}

func testArchive(t *testing.T, r io.Reader) []string {
	// Finish the archiving process in-memory
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("err: %s", err)
	}

	gzipR, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tarR := tar.NewReader(gzipR)

	// Read all the entries
	result := make([]string, 0, 5)
	for {
		hdr, err := tarR.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		result = append(result, hdr.Name)
	}

	sort.Strings(result)
	return result
}
