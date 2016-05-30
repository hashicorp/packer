package lxd

import (
	"os"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"output_image": "foo",
		"image":        "bar",
	}
}

func TestBuilder_Foo(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("This test is only run with PACKER_ACC=1")
	}
}

func TestBuilderPrepare_ConfigFile(t *testing.T) {
	var b Builder
	// Good
	config := testConfig()
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Bad, missing image name
	config = testConfig()
	delete(config, "image")
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should have error")
	}

}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}
