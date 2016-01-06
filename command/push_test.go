package command

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
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
	var actual []string
	var actualOpts *uploadOpts
	uploadFn := func(r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		actual = testArchive(t, r)
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

	expected := []string{
		archiveTemplateEntry,
		"template.json",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}

	expectedBuilds := map[string]*uploadBuildInfo{
		"dummy": &uploadBuildInfo{
			Type: "dummy",
		},
	}
	if !reflect.DeepEqual(actualOpts.Builds, expectedBuilds) {
		t.Fatalf("bad: %#v", actualOpts.Builds)
	}
}

func TestPush_builds(t *testing.T) {
	var actualOpts *uploadOpts
	uploadFn := func(
		r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		actualOpts = opts

		doneCh := make(chan struct{})
		close(doneCh)
		return doneCh, nil, nil
	}

	c := &PushCommand{
		Meta:     testMeta(t),
		uploadFn: uploadFn,
	}

	args := []string{filepath.Join(testFixture("push-builds"), "template.json")}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	expectedBuilds := map[string]*uploadBuildInfo{
		"dummy": &uploadBuildInfo{
			Type:     "dummy",
			Artifact: true,
		},
		"foo": &uploadBuildInfo{
			Type: "dummy",
		},
	}
	if !reflect.DeepEqual(actualOpts.Builds, expectedBuilds) {
		t.Fatalf("bad: %#v", actualOpts.Builds)
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

func TestPush_cliName(t *testing.T) {
	var actual []string
	uploadFn := func(r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		actual = testArchive(t, r)

		doneCh := make(chan struct{})
		close(doneCh)
		return doneCh, nil, nil
	}

	c := &PushCommand{
		Meta:     testMeta(t),
		uploadFn: uploadFn,
	}

	args := []string{
		"-name=foo/bar",
		filepath.Join(testFixture("push-no-name"), "template.json"),
	}

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	expected := []string{
		archiveTemplateEntry,
		"template.json",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
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

func TestPush_vars(t *testing.T) {
	var actualOpts *uploadOpts
	uploadFn := func(r io.Reader, opts *uploadOpts) (<-chan struct{}, <-chan error, error) {
		actualOpts = opts

		doneCh := make(chan struct{})
		close(doneCh)
		return doneCh, nil, nil
	}

	c := &PushCommand{
		Meta:     testMeta(t),
		uploadFn: uploadFn,
	}

	args := []string{
		"-var", "name=foo/bar",
		filepath.Join(testFixture("push-vars"), "template.json"),
	}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	expected := "foo/bar"
	if actualOpts.Slug != expected {
		t.Fatalf("bad: %#v", actualOpts.Slug)
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
