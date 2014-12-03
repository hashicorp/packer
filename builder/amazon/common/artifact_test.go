package common

import (
	"reflect"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}

func TestArtifactId(t *testing.T) {
	expected := `east:foo,west:bar`

	amis := make(map[string]string)
	amis["east"] = "foo"
	amis["west"] = "bar"

	a := &Artifact{
		Amis: amis,
	}

	result := a.Id()
	if result != expected {
		t.Fatalf("bad: %s", result)
	}
}

func TestArtifactState_atlasMetadata(t *testing.T) {
	a := &Artifact{
		Amis: map[string]string{
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

func TestArtifactString(t *testing.T) {
	expected := `AMIs were created:

east: foo
west: bar`

	amis := make(map[string]string)
	amis["east"] = "foo"
	amis["west"] = "bar"

	a := &Artifact{Amis: amis}
	result := a.String()
	if result != expected {
		t.Fatalf("bad: %s", result)
	}
}
