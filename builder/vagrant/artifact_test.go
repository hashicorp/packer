package vagrant

import (
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{} = &artifact{}

	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact does not implement packer.Artifact")
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
	artifact := &artifact{
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
}
