package yandex

import (
	"testing"

	"github.com/hashicorp/packer/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func TestArtifact_impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}

func TestArtifact_Id(t *testing.T) {
	i := &compute.Image{
		Id:       "test-id-value",
		FolderId: "test-folder-id",
	}
	a := &Artifact{
		image: i}
	expected := "test-id-value"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifact_String(t *testing.T) {
	i := &compute.Image{
		Id:       "test-id-value",
		FolderId: "test-folder-id",
		Name:     "test-name",
		Family:   "test-family",
	}
	a := &Artifact{
		image: i}
	expected := "A disk image was created: test-name (id: test-id-value) with family name test-family"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
