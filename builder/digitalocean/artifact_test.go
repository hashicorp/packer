package digitalocean

import (
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func generatedData() map[string]interface{} {
	return make(map[string]interface{})
}

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packersdk.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactId(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo", "tor1"}, nil, generatedData()}
	expected := "sfo,tor1:42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactIdWithoutMultipleRegions(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo"}, nil, generatedData()}
	expected := "sfo:42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo", "tor1"}, nil, generatedData()}
	expected := "A snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo,tor1'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}

func TestArtifactStringWithoutMultipleRegions(t *testing.T) {
	a := &Artifact{"packer-foobar", 42, []string{"sfo"}, nil, generatedData()}
	expected := "A snapshot was created: 'packer-foobar' (ID: 42) in regions 'sfo'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
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
