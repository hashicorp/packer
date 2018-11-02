package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/configfile"
)

func TestArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(artifact)
}

func TestNewArtifact(t *testing.T) {
	prefix, _ := configfile.ConfigTmpDir()
	td, err := ioutil.TempDir(prefix, "hyperv")
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

	a, err := NewArtifact(td)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if a.BuilderId() != BuilderId {
		t.Fatalf("bad: %#v", a.BuilderId())
	}
	if len(a.Files()) != 1 {
		t.Fatalf("should length 1: %d", len(a.Files()))
	}
}
