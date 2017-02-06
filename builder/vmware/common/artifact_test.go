package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestLocalArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(artifact)
}

func TestNewLocalArtifact(t *testing.T) {
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	err = ioutil.WriteFile(filepath.Join(td, "a"), []byte("foo"), 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := os.Mkdir(filepath.Join(td, "b"), 0755); err != nil {
		t.Fatalf("err: %s", err)
	}
	dir := new(LocalOutputDir)
	dir.SetOutputDir(td)
	files, err := dir.ListFiles()

	a := NewArtifact(dir, files, false)

	if a.BuilderId() != BuilderId {
		t.Fatalf("bad: %#v", a.BuilderId())
	}
	if len(a.Files()) != 1 {
		t.Fatalf("should length 1: %d", len(a.Files()))
	}
}
