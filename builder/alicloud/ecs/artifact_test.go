package ecs

import (
	"reflect"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifactId(t *testing.T) {
	expected := `east:foo,west:bar`

	ecsImages := make(map[string]string)
	ecsImages["east"] = "foo"
	ecsImages["west"] = "bar"

	a := &Artifact{
		AlicloudImages: ecsImages,
	}

	result := a.Id()
	if result != expected {
		t.Fatalf("bad: %s", result)
	}
}

func TestArtifactState_atlasMetadata(t *testing.T) {
	a := &Artifact{
		AlicloudImages: map[string]string{
			"east": "foo",
			"west": "bar",
		},
	}

	actual := a.State("atlas.artifact.metadata")
	expected := map[string]string{
		"region.east": "foo",
		"region.west": "bar",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}
