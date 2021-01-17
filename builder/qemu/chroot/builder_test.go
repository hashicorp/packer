package chroot

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"source_image": "foo",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilderPrepare_SourceImage(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["source_image"] = "foo"
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	delete(config, "source_image")
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ChrootMounts(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["chroot_mounts"] = nil
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
	}
}

func TestBuilderPrepare_ChrootMountsBadDefaults(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["chroot_mounts"] = [][]string{
		{"bad"},
	}
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_CommandWrapper(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["command_wrapper"] = "echo hi; {{.Command}}"
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
	}
}

func TestBuilderPrepare_CopyFiles(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
	}

	if len(b.config.CopyFiles) != 1 && b.config.CopyFiles[0] != "/etc/resolv.conf" {
		t.Errorf("Was expecting default value for copy_files.")
	}
}
func TestBuilderPrepare_CopyFilesNoDefault(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["copy_files"] = []string{}
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
	}

	if len(b.config.CopyFiles) > 0 {
		t.Errorf("Was expecting no default value for copy_files. Found %v",
			b.config.CopyFiles)
	}
}
func TestBuilderPrepare_ReturnGeneratedData(t *testing.T) {
	var b Builder
	config := testConfig()

	generatedData, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if generatedData[0] != "Device" {
		t.Fatalf("Generated data should contain Device")
	}
	if generatedData[1] != "MountPath" {
		t.Fatalf("Generated data should contain MountPath")
	}
}
