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
west: bar
`

	amis := make(map[string]string)
	amis["east"] = "foo"
	amis["west"] = "bar"

	a := &Artifact{Amis: amis}
	result := a.String()
	if result != expected {
		t.Fatalf("bad: %s", result)
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
