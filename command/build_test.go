package command

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/packer/builder/file"
	"github.com/mitchellh/packer/packer"
)

func TestBuildOnlyFileCommaFlags(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-only=chocolate,vanilla",
		filepath.Join(testFixture("build-only"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	if !fileExists("chocolate.txt") {
		t.Error("Expected to find chocolate.txt")
	}
	if !fileExists("vanilla.txt") {
		t.Error("Expected to find vanilla.txt")
	}
	if fileExists("cherry.txt") {
		t.Error("Expected NOT to find cherry.txt")
	}
}

func TestBuildStdin(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}
	f, err := os.Open(filepath.Join(testFixture("build-only"), "template.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stdin := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = stdin }()

	defer cleanup()
	if code := c.Run([]string{"-"}); code != 0 {
		fatalCommand(t, c.Meta)
	}

	if !fileExists("chocolate.txt") {
		t.Error("Expected to find chocolate.txt")
	}
	if !fileExists("vanilla.txt") {
		t.Error("Expected to find vanilla.txt")
	}
	if !fileExists("cherry.txt") {
		t.Error("Expected to find cherry.txt")
	}
}

func TestBuildOnlyFileMultipleFlags(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-only=chocolate",
		"-only=cherry",
		filepath.Join(testFixture("build-only"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	if !fileExists("chocolate.txt") {
		t.Error("Expected to find chocolate.txt")
	}
	if fileExists("vanilla.txt") {
		t.Error("Expected NOT to find vanilla.txt")
	}
	if !fileExists("cherry.txt") {
		t.Error("Expected to find cherry.txt")
	}
}

func TestBuildExceptFileCommaFlags(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-except=chocolate",
		filepath.Join(testFixture("build-only"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	if fileExists("chocolate.txt") {
		t.Error("Expected NOT to find chocolate.txt")
	}
	if !fileExists("vanilla.txt") {
		t.Error("Expected to find vanilla.txt")
	}
	if !fileExists("cherry.txt") {
		t.Error("Expected to find cherry.txt")
	}
}

// fileExists returns true if the filename is found
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// testCoreConfigBuilder creates a packer CoreConfig that has a file builder
// available. This allows us to test a builder that writes files to disk.
func testCoreConfigBuilder(t *testing.T) *packer.CoreConfig {
	components := packer.ComponentFinder{
		Builder: func(n string) (packer.Builder, error) {
			return &file.Builder{}, nil
		},
	}
	return &packer.CoreConfig{
		Components: components,
	}
}

// testMetaFile creates a Meta object that includes a file builder
func testMetaFile(t *testing.T) Meta {
	var out, err bytes.Buffer
	return Meta{
		CoreConfig: testCoreConfigBuilder(t),
		Ui: &packer.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func cleanup() {
	os.RemoveAll("chocolate.txt")
	os.RemoveAll("vanilla.txt")
	os.RemoveAll("cherry.txt")
}
