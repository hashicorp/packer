package common

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
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
