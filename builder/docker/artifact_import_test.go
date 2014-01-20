package docker

import (
	"errors"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestImportArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(ImportArtifact)
}

func TestImportArtifactBuilderId(t *testing.T) {
	a := &ImportArtifact{BuilderIdValue: "foo"}
	if a.BuilderId() != "foo" {
		t.Fatalf("bad: %#v", a.BuilderId())
	}
}

func TestImportArtifactFiles(t *testing.T) {
	a := &ImportArtifact{}
	if a.Files() != nil {
		t.Fatalf("bad: %#v", a.Files())
	}
}

func TestImportArtifactId(t *testing.T) {
	a := &ImportArtifact{IdValue: "foo"}
	if a.Id() != "foo" {
		t.Fatalf("bad: %#v", a.Id())
	}
}

func TestImportArtifactDestroy(t *testing.T) {
	d := new(MockDriver)
	a := &ImportArtifact{
		Driver:  d,
		IdValue: "foo",
	}

	// No error
	if err := a.Destroy(); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !d.DeleteImageCalled {
		t.Fatal("delete image should be called")
	}
	if d.DeleteImageId != "foo" {
		t.Fatalf("bad: %#v", d.DeleteImageId)
	}

	// With an error
	d.DeleteImageErr = errors.New("foo")
	if err := a.Destroy(); err != d.DeleteImageErr {
		t.Fatalf("err: %#v", err)
	}
}
