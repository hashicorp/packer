package chroot

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"ami_name":   "foo",
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

func TestBuilderPrepare_AMIName(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["ami_name"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ami_name"] = "foo {{"
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad
	delete(config, "ami_name")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_Description(t *testing.T) {
	var b Builder
	config := testConfig()
	
	// Test with no description field
	config["description"] = ""
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if b.config.Description != "<No Description>" {
		t.Fatalf("Wrong value: %s", b.config.Description)
	}

	// Test with a description field
	config["description"] = "VM description"
	err = b.Prepare(config)
	if b.config.Description != "VM description" {
		t.Fatalf("Wrong value: %s", b.config.Description)
	}
}

func TestBuilderPrepare_ChrootMounts(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["chroot_mounts"] = nil
	err := b.Prepare(config)
	if err != nil {
		t.Errorf("err: %s", err)
	}

	config["chroot_mounts"] = [][]string{
		[]string{"bad"},
	}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
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

func TestBuilderPrepare_CommandWrapper(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["command_wrapper"] = "echo hi; {{.Command}}"
	err := b.Prepare(config)
	if err != nil {
		t.Errorf("err: %s", err)
	}
}
