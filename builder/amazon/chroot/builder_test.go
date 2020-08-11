package chroot

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"ami_name":   "foo",
		"source_ami": "foo",
		"region":     "us-east-1",
		// region validation logic is checked in ami_config_test
		"skip_region_validation": true,
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
	config["skip_region_validation"] = true
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ami_name"] = "foo {{"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad
	delete(config, "ami_name")
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
func TestBuilderPrepare_SourceAmi(t *testing.T) {
	b := &Builder{}
	config := testConfig()

	config["source_ami"] = ""
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["source_ami"] = "foo"
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
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

func TestBuilderPrepare_RootDeviceNameAndAMIMappings(t *testing.T) {
	var b Builder
	config := testConfig()

	config["root_device_name"] = "/dev/sda"
	config["ami_block_device_mappings"] = []interface{}{map[string]string{}}
	config["root_volume_size"] = 15
	_, warnings, err := b.Prepare(config)
	if len(warnings) == 0 {
		t.Fatal("Missing warning, stating block device mappings will be overwritten")
	} else if len(warnings) > 1 {
		t.Fatalf("excessive warnings: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_AMIMappingsNoRootDeviceName(t *testing.T) {
	var b Builder
	config := testConfig()

	config["ami_block_device_mappings"] = []interface{}{map[string]string{}}
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should have error")
	}
}

func TestBuilderPrepare_RootDeviceNameNoAMIMappings(t *testing.T) {
	var b Builder
	config := testConfig()

	config["root_device_name"] = "/dev/sda"
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should have error")
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
	if len(generatedData) == 0 {
		t.Fatalf("Generated data should not be empty")
	}
	if generatedData[0] != "SourceAMIName" {
		t.Fatalf("Generated data should contain SourceAMIName")
	}
	if generatedData[1] != "BuildRegion" {
		t.Fatalf("Generated data should contain BuildRegion")
	}
	if generatedData[2] != "SourceAMI" {
		t.Fatalf("Generated data should contain SourceAMI")
	}
	if generatedData[3] != "SourceAMICreationDate" {
		t.Fatalf("Generated data should contain SourceAMICreationDate")
	}
	if generatedData[4] != "SourceAMIOwner" {
		t.Fatalf("Generated data should contain SourceAMIOwner")
	}
	if generatedData[5] != "SourceAMIOwnerName" {
		t.Fatalf("Generated data should contain SourceAMIOwnerName")
	}
	if generatedData[6] != "Device" {
		t.Fatalf("Generated data should contain Device")
	}
	if generatedData[7] != "MountPath" {
		t.Fatalf("Generated data should contain MountPath")
	}
}
