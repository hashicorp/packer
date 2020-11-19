package yandex

import (
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func TestArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifact_Id(t *testing.T) {
	i := &compute.Image{
		Id:       "test-id-value",
		FolderId: "test-folder-id",
	}
	a := &Artifact{
		Image: i}
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
		Image: i}
	expected := "A disk image was created: test-name (id: test-id-value) with family name test-family"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}

func TestArtifactState(t *testing.T) {
	expectedData := "this is the data"
	artifact := &Artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := artifact.State("state_data")
	if result != expectedData {
		t.Fatalf("Bad: State data was %s instead of %s", result, expectedData)
	}

	// Invalid state
	result = artifact.State("invalid_key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for invalid state data name")
	}

	// Nil StateData should not fail and should return nil
	artifact = &Artifact{}
	result = artifact.State("key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for nil StateData")
	}
}
