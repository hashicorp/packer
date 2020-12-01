package vagrant

import (
	"runtime"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{} = &artifact{}

	if _, ok := raw.(packersdk.Artifact); !ok {
		t.Fatalf("Artifact does not implement packersdk.Artifact")
	}
}

func TestArtifactId(t *testing.T) {
	a := &artifact{
		OutputDir: "/my/dir",
		BoxName:   "package.box",
		Provider:  "virtualbox",
	}

	expected := "virtualbox"
	if a.Id() != expected {
		t.Fatalf("artifact ID should match: expected: %s received: %s", expected, a.Id())
	}
}

func TestArtifactString(t *testing.T) {
	a := &artifact{
		OutputDir: "/my/dir",
		BoxName:   "package.box",
		Provider:  "virtualbox",
	}
	expected := "Vagrant box 'package.box' for 'virtualbox' provider"
	if runtime.GOOS == "windows" {
		expected = strings.Replace(expected, "/", "\\", -1)
	}

	if strings.Compare(a.String(), expected) != 0 {
		t.Fatalf("artifact string should match: expected: %s received: %s", expected, a.String())
	}
}

func TestArtifactState(t *testing.T) {
	expectedData := "this is the data"
	a := &artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := a.State("state_data")
	if result != expectedData {
		t.Fatalf("Bad: State data was %s instead of %s", result, expectedData)
	}

	// Invalid state
	result = a.State("invalid_key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for invalid state data name")
	}

	// Nil StateData should not fail and should return nil
	a = &artifact{}
	result = a.State("key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for nil StateData")
	}
}
