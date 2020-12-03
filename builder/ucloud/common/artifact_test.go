package common

import (
	"reflect"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifactId(t *testing.T) {
	expected := `project1:region1:foo,project2:region2:bar`

	images := NewImageInfoSet(nil)
	images.Set(ImageInfo{
		Region:    "region1",
		ProjectId: "project1",
		ImageId:   "foo",
	})

	images.Set(ImageInfo{
		Region:    "region2",
		ProjectId: "project2",
		ImageId:   "bar",
	})

	a := &Artifact{
		UCloudImages: images,
	}

	result := a.Id()
	if result != expected {
		t.Fatalf("bad: %s", result)
	}
}

func TestArtifactState_atlasMetadata(t *testing.T) {
	images := NewImageInfoSet(nil)
	images.Set(ImageInfo{
		Region:    "region1",
		ProjectId: "project1",
		ImageId:   "foo",
	})

	images.Set(ImageInfo{
		Region:    "region2",
		ProjectId: "project2",
		ImageId:   "bar",
	})

	a := &Artifact{
		UCloudImages: images,
	}

	actual := a.State("atlas.artifact.metadata")
	expected := map[string]string{
		"project1:region1": "foo",
		"project2:region2": "bar",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestArtifactState_StateData(t *testing.T) {
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
