package chroot

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"source_ami": "foo",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilderPrepare_SourceAmi(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["source_ami"] = ""
	err := b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["source_ami"] = "foo"
	err = b.Prepare(config)
	if err != nil {
		t.Errorf("err: %s", err)
	}
}
