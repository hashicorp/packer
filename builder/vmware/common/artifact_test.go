package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestLocalArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(localArtifact)
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

	a, err := NewLocalArtifact("vm1", td)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if a.BuilderId() != BuilderId {
		t.Fatalf("bad: %#v", a.BuilderId())
	}
	if a.Id() != "vm1" {
		t.Fatalf("bad: %#v", a.Id())
	}
	if len(a.Files()) != 1 {
		t.Fatalf("should length 1: %d", len(a.Files()))
	}
}
