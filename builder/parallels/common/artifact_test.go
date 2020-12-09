package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(artifact)
}

func TestNewArtifact(t *testing.T) {
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	err = ioutil.WriteFile(filepath.Join(td, "a"), []byte("foo"), 0644)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err = os.Mkdir(filepath.Join(td, "b"), 0755); err != nil {
		t.Fatalf("err: %s", err)
	}

	generatedData := map[string]interface{}{"generated_data": "data"}
	a, err := NewArtifact(td, generatedData)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if a.BuilderId() != BuilderId {
		t.Fatalf("bad: %#v", a.BuilderId())
	}
	if len(a.Files()) != 1 {
		t.Fatalf("should length 1: %d", len(a.Files()))
	}
	if a.State("generated_data") != "data" {
		t.Fatalf("bad: should length have generated_data: %s", a.State("generated_data"))
	}
}
