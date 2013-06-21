package docker

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"repository": "foo/bar",
		"build_path": "../..",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilderPrepare_Repository(t *testing.T) {
	var b Builder
	config := testConfig()

	if err := b.Prepare(config); err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test blank
	config["repository"] = ""
	if err := b.Prepare(config); err == nil {
		t.Fatalf("should have error: %s", err)
	}
}

func TestBuilderPrepare_BuildPath(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	b = Builder{}
	if err := b.Prepare(config); err != nil {
		t.Fatal("should not have error")
	}

	// Test default
	config["build_path"] = ""
	if err := b.Prepare(config); err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.BuildPath != "." {
		t.Errorf("Did not set up default BuildPath")
	}
}
